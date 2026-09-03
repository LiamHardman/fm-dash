package main

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"sort"
	"sync"

	duckdb "github.com/duckdb/duckdb-go/v2"
)

const createCategoryWeightsTableDDL = `
CREATE TABLE IF NOT EXISTS category_weights (
    category  VARCHAR NOT NULL,
    attribute VARCHAR NOT NULL,
    weight    INTEGER NOT NULL,
    PRIMARY KEY (category, attribute)
);`

const createRoleWeightsTableDDL = `
CREATE TABLE IF NOT EXISTS role_weights (
    role_name      VARCHAR NOT NULL,
    position       VARCHAR NOT NULL,
    style_category VARCHAR NOT NULL,
    attribute      VARCHAR NOT NULL,
    weight         INTEGER NOT NULL,
    PRIMARY KEY (role_name, attribute)
);`

const createCasWeightsTableDDL = `
CREATE TABLE IF NOT EXISTS cas_weights (
    position  VARCHAR NOT NULL,
    attribute VARCHAR NOT NULL,
    weight    DOUBLE NOT NULL,
    PRIMARY KEY (position, attribute)
);`

// maxAttributeWeightValue mirrors src/api/config.go's identical constant --
// keep the two in sync if it's ever changed there.
const maxAttributeWeightValue = 1000

// WeightsValidationError marks a client-input validation failure (HTTP 400),
// as distinct from an internal/database error (HTTP 500).
type WeightsValidationError struct{ msg string }

func (e *WeightsValidationError) Error() string { return e.msg }

// WeightsStore owns the single persistent DuckDB connection holding
// attribute weights, at <DatasetStorageDir>/weights.duckdb. Reads use the
// standard database/sql pool (DuckDB's MVCC handles concurrent readers
// safely within one process, and this service is a single-instance process
// per the Dataset Query Service's standing design). Writes are additionally
// serialized by mu purely to avoid ever needing MVCC-conflict-retry logic
// for what is a low-frequency admin path.
type WeightsStore struct {
	db *sql.DB
	mu sync.Mutex
}

func openWeightsStore(ctx context.Context, storageDir string) (*WeightsStore, error) {
	dbPath := filepath.Join(storageDir, "weights.duckdb")
	connector, err := duckdb.NewConnector(dbPath, nil)
	if err != nil {
		return nil, fmt.Errorf("creating duckdb connector for %s: %w", dbPath, err)
	}
	db := sql.OpenDB(connector)

	if _, err := db.ExecContext(ctx, createCategoryWeightsTableDDL); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("creating category_weights table: %w", err)
	}
	if _, err := db.ExecContext(ctx, createRoleWeightsTableDDL); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("creating role_weights table: %w", err)
	}
	if _, err := db.ExecContext(ctx, createCasWeightsTableDDL); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("creating cas_weights table: %w", err)
	}
	return &WeightsStore{db: db}, nil
}

// Close closes the store's underlying DuckDB connection pool.
func (s *WeightsStore) Close() error { return s.db.Close() }

// GetAll returns every category's weights as map[string]map[string]int --
// the same shape SetCategories accepts, for a symmetric read/write contract.
func (s *WeightsStore) GetAll(ctx context.Context) (map[string]map[string]int, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT category, attribute, weight FROM category_weights ORDER BY category, attribute`)
	if err != nil {
		return nil, fmt.Errorf("querying category_weights: %w", err)
	}
	defer func() { _ = rows.Close() }()

	result := make(map[string]map[string]int)
	for rows.Next() {
		var category, attribute string
		var weight int
		if err := rows.Scan(&category, &attribute, &weight); err != nil {
			return nil, fmt.Errorf("scanning weight row: %w", err)
		}
		if result[category] == nil {
			result[category] = make(map[string]int)
		}
		result[category][attribute] = weight
	}
	return result, rows.Err()
}

// validateWeightsUpdate mirrors src/api/config.go's SetAttributeWeights
// validation exactly: update must be non-empty, each category's weight map
// must be non-empty, every weight in [0, maxAttributeWeightValue].
func validateWeightsUpdate(update map[string]map[string]int) error {
	if len(update) == 0 {
		return &WeightsValidationError{"weights update must not be empty"}
	}
	for category, weights := range update {
		if len(weights) == 0 {
			return &WeightsValidationError{fmt.Sprintf("category %q has no attribute weights", category)}
		}
		for attr, w := range weights {
			if w < 0 || w > maxAttributeWeightValue {
				return &WeightsValidationError{fmt.Sprintf(
					"category %q: weight for %q must be between 0 and %d", category, attr, maxAttributeWeightValue)}
			}
		}
	}
	return nil
}

// SetCategories upserts update into category_weights with category-replace
// semantics: for each category present in update, ALL of that category's
// existing rows are deleted and replaced with exactly the rows in update;
// categories not mentioned in update are left completely untouched -- this
// is the exact behavior of src/api/config.go's SetAttributeWeights. All
// categories in one call commit atomically together.
func (s *WeightsStore) SetCategories(ctx context.Context, update map[string]map[string]int) error {
	if err := validateWeightsUpdate(update); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning weights transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // no-op once committed

	categories := make([]string, 0, len(update))
	for c := range update {
		categories = append(categories, c)
	}
	sort.Strings(categories) // deterministic order for debuggability only

	del, err := tx.PrepareContext(ctx, `DELETE FROM category_weights WHERE category = ?`)
	if err != nil {
		return fmt.Errorf("preparing delete: %w", err)
	}
	defer func() { _ = del.Close() }()

	ins, err := tx.PrepareContext(ctx, `INSERT INTO category_weights (category, attribute, weight) VALUES (?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("preparing insert: %w", err)
	}
	defer func() { _ = ins.Close() }()

	for _, category := range categories {
		if _, err := del.ExecContext(ctx, category); err != nil {
			return fmt.Errorf("deleting existing weights for category %q: %w", category, err)
		}
		for attr, w := range update[category] {
			if _, err := ins.ExecContext(ctx, category, attr, w); err != nil {
				return fmt.Errorf("inserting weight %s.%s: %w", category, attr, err)
			}
		}
	}

	return tx.Commit()
}

// seedDefaultsIfEmpty copies defaultAttributeWeights (weights_defaults.go)
// into category_weights the first time the store is opened with no rows
// present -- "stored inside fmdash at the start, then copied to duckdb."
// Subsequent restarts leave whatever is already in DuckDB untouched,
// including any admin edits made via PUT /internal/weights.
func (s *WeightsStore) seedDefaultsIfEmpty(ctx context.Context) error {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM category_weights`).Scan(&count); err != nil {
		return fmt.Errorf("counting category_weights: %w", err)
	}
	if count > 0 {
		return nil
	}
	return s.SetCategories(ctx, defaultAttributeWeights)
}

// seedRoleWeightsIfEmpty copies the embedded role_weights.json (a verbatim
// mirror of src/api/public/role_specific_overall_weights.json) into
// role_weights the first time the store is opened with no rows present.
// position is precomputed as the text before the role name's FIRST " - "
// (roleStylePosition); style_category as the text before the LAST " - "
// (roleStyleCategory) -- both computed once here in Go from the same logic
// src/api uses, rather than re-derived per query in SQL. Unlike category
// weights, there is no admin write endpoint for role weights in this
// phase (src/api has no runtime role-weight-editing feature either), so
// this is seed-once, not seed-if-empty-then-editable.
func (s *WeightsStore) seedRoleWeightsIfEmpty(ctx context.Context) error {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM role_weights`).Scan(&count); err != nil {
		return fmt.Errorf("counting role_weights: %w", err)
	}
	if count > 0 {
		return nil
	}

	roleWeights, err := parseEmbeddedRoleWeights()
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning role weights transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	ins, err := tx.PrepareContext(ctx,
		`INSERT INTO role_weights (role_name, position, style_category, attribute, weight) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("preparing role weights insert: %w", err)
	}
	defer func() { _ = ins.Close() }()

	roleNames := make([]string, 0, len(roleWeights))
	for roleName := range roleWeights {
		roleNames = append(roleNames, roleName)
	}
	sort.Strings(roleNames) // deterministic insert order for debuggability only

	for _, roleName := range roleNames {
		position := roleStylePosition(roleName)
		styleCategory := roleStyleCategory(roleName)
		for attr, weight := range roleWeights[roleName] {
			if _, err := ins.ExecContext(ctx, roleName, position, styleCategory, attr, weight); err != nil {
				return fmt.Errorf("inserting role weight %s.%s: %w", roleName, attr, err)
			}
		}
	}

	return tx.Commit()
}

// seedCasWeightsIfEmpty copies the embedded cas_weights.json (extracted from
// src/api/ca_calculation.go's casPositionWeights/casAttrOrder Go literal via
// a one-off, never-committed dump program -- see cas_weights_seed.go) into
// cas_weights the first time the store is opened with no rows present.
// Zero-weight (position, attribute) pairs are already omitted from the
// embedded JSON, matching calculateCASForPosition's exclude-zero-weight
// behavior. Like role weights, there is no admin write endpoint for CAS
// weights in this phase, so this is seed-once, not seed-if-empty-then-editable.
func (s *WeightsStore) seedCasWeightsIfEmpty(ctx context.Context) error {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM cas_weights`).Scan(&count); err != nil {
		return fmt.Errorf("counting cas_weights: %w", err)
	}
	if count > 0 {
		return nil
	}

	casWeights, err := parseEmbeddedCasWeights()
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning cas weights transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	ins, err := tx.PrepareContext(ctx,
		`INSERT INTO cas_weights (position, attribute, weight) VALUES (?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("preparing cas weights insert: %w", err)
	}
	defer func() { _ = ins.Close() }()

	positions := make([]string, 0, len(casWeights))
	for position := range casWeights {
		positions = append(positions, position)
	}
	sort.Strings(positions) // deterministic insert order for debuggability only

	for _, position := range positions {
		for attr, weight := range casWeights[position] {
			if _, err := ins.ExecContext(ctx, position, attr, weight); err != nil {
				return fmt.Errorf("inserting cas weight %s.%s: %w", position, attr, err)
			}
		}
	}

	return tx.Commit()
}
