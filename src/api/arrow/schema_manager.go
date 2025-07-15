package arrow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/apache/arrow/go/v17/arrow"
)

// SchemaManager interface defines operations for schema management and evolution
type SchemaManager interface {
	GetCurrentSchema() *arrow.Schema
	GetSchemaByVersion(version int) (*arrow.Schema, error)
	ValidateSchema(schema *arrow.Schema) error
	EvolveSchema(oldSchema, newSchema *arrow.Schema) (*SchemaEvolution, error)
	RegisterSchema(version int, schema *arrow.Schema) error
	GetSchemaHistory() ([]SchemaVersion, error)
	PlanMigration(fromVersion, toVersion int) (*MigrationPlan, error)
}

// SchemaManagerImpl implements the SchemaManager interface
type SchemaManagerImpl struct {
	schemaDir      string
	currentVersion int
	schemas        map[int]*arrow.Schema
	validator      *SchemaValidator
}

// SchemaVersion represents a schema version with metadata
type SchemaVersion struct {
	Version     int                    `json:"version"`
	Schema      *SerializableSchema    `json:"schema"`
	CreatedAt   time.Time             `json:"created_at"`
	Description string                `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SerializableSchema represents a schema that can be serialized to JSON
type SerializableSchema struct {
	Fields []SerializableField `json:"fields"`
}

// SerializableField represents a field that can be serialized to JSON
type SerializableField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
}

// SchemaEvolution represents the result of schema evolution analysis
type SchemaEvolution struct {
	FromVersion    int                  `json:"from_version"`
	ToVersion      int                  `json:"to_version"`
	AddedFields    []arrow.Field        `json:"added_fields"`
	RemovedFields  []arrow.Field        `json:"removed_fields"`
	ModifiedFields []FieldModification  `json:"modified_fields"`
	Compatible     bool                 `json:"compatible"`
	MigrationPlan  []MigrationStep      `json:"migration_plan"`
	CreatedAt      time.Time            `json:"created_at"`
}

// FieldModification represents a modification to a field
type FieldModification struct {
	FieldName   string      `json:"field_name"`
	OldType     string      `json:"old_type"`
	NewType     string      `json:"new_type"`
	OldNullable bool        `json:"old_nullable"`
	NewNullable bool        `json:"new_nullable"`
	Compatible  bool        `json:"compatible"`
	Reason      string      `json:"reason"`
}

// MigrationStep represents a single step in a migration plan
type MigrationStep struct {
	StepType    MigrationStepType `json:"step_type"`
	Description string            `json:"description"`
	FieldName   string            `json:"field_name,omitempty"`
	OldType     string            `json:"old_type,omitempty"`
	NewType     string            `json:"new_type,omitempty"`
	Required    bool              `json:"required"`
	Reversible  bool              `json:"reversible"`
}

// MigrationStepType represents the type of migration step
type MigrationStepType int

const (
	MigrationStepAddField MigrationStepType = iota
	MigrationStepRemoveField
	MigrationStepModifyField
	MigrationStepValidateData
	MigrationStepBackupData
	MigrationStepTransformData
)

func (mst MigrationStepType) String() string {
	switch mst {
	case MigrationStepAddField:
		return "ADD_FIELD"
	case MigrationStepRemoveField:
		return "REMOVE_FIELD"
	case MigrationStepModifyField:
		return "MODIFY_FIELD"
	case MigrationStepValidateData:
		return "VALIDATE_DATA"
	case MigrationStepBackupData:
		return "BACKUP_DATA"
	case MigrationStepTransformData:
		return "TRANSFORM_DATA"
	default:
		return "UNKNOWN"
	}
}

// MigrationPlan represents a complete migration plan between schema versions
type MigrationPlan struct {
	FromVersion int               `json:"from_version"`
	ToVersion   int               `json:"to_version"`
	Steps       []MigrationStep   `json:"steps"`
	Reversible  bool              `json:"reversible"`
	EstimatedDuration time.Duration `json:"estimated_duration"`
	CreatedAt   time.Time         `json:"created_at"`
}

// NewSchemaManager creates a new SchemaManager instance
func NewSchemaManager(schemaDir string) (*SchemaManagerImpl, error) {
	if schemaDir == "" {
		schemaDir = "./schemas"
	}

	// Create schema directory if it doesn't exist
	if err := os.MkdirAll(schemaDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create schema directory: %w", err)
	}

	sm := &SchemaManagerImpl{
		schemaDir:      schemaDir,
		currentVersion: GetSchemaVersion(),
		schemas:        make(map[int]*arrow.Schema),
		validator:      NewSchemaValidator(),
	}

	// Load existing schemas
	if err := sm.loadSchemas(); err != nil {
		return nil, fmt.Errorf("failed to load schemas: %w", err)
	}

	// Register current schema if not already registered
	if _, exists := sm.schemas[sm.currentVersion]; !exists {
		if err := sm.RegisterSchema(sm.currentVersion, PlayerSchema); err != nil {
			return nil, fmt.Errorf("failed to register current schema: %w", err)
		}
	}

	return sm, nil
}

// GetCurrentSchema returns the current schema
func (sm *SchemaManagerImpl) GetCurrentSchema() *arrow.Schema {
	return PlayerSchema
}

// GetSchemaByVersion returns a schema by version number
func (sm *SchemaManagerImpl) GetSchemaByVersion(version int) (*arrow.Schema, error) {
	schema, exists := sm.schemas[version]
	if !exists {
		return nil, fmt.Errorf("schema version %d not found", version)
	}
	return schema, nil
}

// ValidateSchema validates a schema
func (sm *SchemaManagerImpl) ValidateSchema(schema *arrow.Schema) error {
	return sm.validator.ValidateSchema(schema)
}

// EvolveSchema analyzes the evolution between two schemas
func (sm *SchemaManagerImpl) EvolveSchema(oldSchema, newSchema *arrow.Schema) (*SchemaEvolution, error) {
	if oldSchema == nil || newSchema == nil {
		return nil, fmt.Errorf("schemas cannot be nil")
	}

	evolution := &SchemaEvolution{
		FromVersion:    sm.getSchemaVersion(oldSchema),
		ToVersion:      sm.getSchemaVersion(newSchema),
		AddedFields:    make([]arrow.Field, 0),
		RemovedFields:  make([]arrow.Field, 0),
		ModifiedFields: make([]FieldModification, 0),
		Compatible:     true,
		MigrationPlan:  make([]MigrationStep, 0),
		CreatedAt:      time.Now(),
	}

	// Compare schemas using existing comparison logic
	comparison := CompareSchemas(oldSchema, newSchema)
	evolution.Compatible = comparison.Compatible

	// Convert schema changes to evolution format
	oldFields := make(map[string]arrow.Field)
	for i, field := range oldSchema.Fields() {
		oldFields[field.Name] = oldSchema.Field(i)
	}

	newFields := make(map[string]arrow.Field)
	for i, field := range newSchema.Fields() {
		newFields[field.Name] = newSchema.Field(i)
	}

	// Identify added fields
	for fieldName, newField := range newFields {
		if _, exists := oldFields[fieldName]; !exists {
			evolution.AddedFields = append(evolution.AddedFields, newField)
			evolution.MigrationPlan = append(evolution.MigrationPlan, MigrationStep{
				StepType:    MigrationStepAddField,
				Description: fmt.Sprintf("Add field '%s' of type %s", fieldName, newField.Type.String()),
				FieldName:   fieldName,
				NewType:     newField.Type.String(),
				Required:    !newField.Nullable,
				Reversible:  true,
			})
		}
	}

	// Identify removed fields
	for fieldName, oldField := range oldFields {
		if _, exists := newFields[fieldName]; !exists {
			evolution.RemovedFields = append(evolution.RemovedFields, oldField)
			evolution.MigrationPlan = append(evolution.MigrationPlan, MigrationStep{
				StepType:    MigrationStepRemoveField,
				Description: fmt.Sprintf("Remove field '%s'", fieldName),
				FieldName:   fieldName,
				OldType:     oldField.Type.String(),
				Required:    true,
				Reversible:  false, // Data loss
			})
		}
	}

	// Identify modified fields
	for fieldName, newField := range newFields {
		if oldField, exists := oldFields[fieldName]; exists {
			if oldField.Type.ID() != newField.Type.ID() || oldField.Nullable != newField.Nullable {
				compatible := sm.validator.areTypesCompatible(oldField.Type, newField.Type)
				if oldField.Nullable && !newField.Nullable {
					compatible = false
				}

				modification := FieldModification{
					FieldName:   fieldName,
					OldType:     oldField.Type.String(),
					NewType:     newField.Type.String(),
					OldNullable: oldField.Nullable,
					NewNullable: newField.Nullable,
					Compatible:  compatible,
				}

				if !compatible {
					evolution.Compatible = false
					modification.Reason = "Incompatible type change or nullability constraint"
				}

				evolution.ModifiedFields = append(evolution.ModifiedFields, modification)
				evolution.MigrationPlan = append(evolution.MigrationPlan, MigrationStep{
					StepType:    MigrationStepModifyField,
					Description: fmt.Sprintf("Modify field '%s' from %s to %s", fieldName, oldField.Type.String(), newField.Type.String()),
					FieldName:   fieldName,
					OldType:     oldField.Type.String(),
					NewType:     newField.Type.String(),
					Required:    true,
					Reversible:  compatible,
				})
			}
		}
	}

	return evolution, nil
}

// RegisterSchema registers a new schema version
func (sm *SchemaManagerImpl) RegisterSchema(version int, schema *arrow.Schema) error {
	if schema == nil {
		return fmt.Errorf("schema cannot be nil")
	}

	// Validate the schema
	if err := sm.ValidateSchema(schema); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	// Convert to serializable format
	serializableSchema := sm.schemaToSerializable(schema)

	schemaVersion := SchemaVersion{
		Version:     version,
		Schema:      serializableSchema,
		CreatedAt:   time.Now(),
		Description: fmt.Sprintf("Schema version %d", version),
		Metadata:    make(map[string]interface{}),
	}

	// Save to file
	filename := filepath.Join(sm.schemaDir, fmt.Sprintf("v%d.json", version))
	data, err := json.MarshalIndent(schemaVersion, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write schema file: %w", err)
	}

	// Store in memory
	sm.schemas[version] = schema

	return nil
}

// GetSchemaHistory returns the history of all schema versions
func (sm *SchemaManagerImpl) GetSchemaHistory() ([]SchemaVersion, error) {
	var history []SchemaVersion

	files, err := filepath.Glob(filepath.Join(sm.schemaDir, "v*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to list schema files: %w", err)
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue // Skip files that can't be read
		}

		var schemaVersion SchemaVersion
		if err := json.Unmarshal(data, &schemaVersion); err != nil {
			continue // Skip files that can't be parsed
		}

		history = append(history, schemaVersion)
	}

	// Sort by version
	sort.Slice(history, func(i, j int) bool {
		return history[i].Version < history[j].Version
	})

	return history, nil
}

// PlanMigration creates a migration plan between two schema versions
func (sm *SchemaManagerImpl) PlanMigration(fromVersion, toVersion int) (*MigrationPlan, error) {
	fromSchema, err := sm.GetSchemaByVersion(fromVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get source schema: %w", err)
	}

	toSchema, err := sm.GetSchemaByVersion(toVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to get target schema: %w", err)
	}

	evolution, err := sm.EvolveSchema(fromSchema, toSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze schema evolution: %w", err)
	}

	plan := &MigrationPlan{
		FromVersion: fromVersion,
		ToVersion:   toVersion,
		Steps:       evolution.MigrationPlan,
		Reversible:  evolution.Compatible,
		CreatedAt:   time.Now(),
	}

	// Estimate duration based on number of steps
	plan.EstimatedDuration = time.Duration(len(plan.Steps)) * time.Minute

	// Add validation and backup steps
	if len(plan.Steps) > 0 {
		// Add backup step at the beginning
		backupStep := MigrationStep{
			StepType:    MigrationStepBackupData,
			Description: "Create backup of existing data",
			Required:    true,
			Reversible:  true,
		}
		plan.Steps = append([]MigrationStep{backupStep}, plan.Steps...)

		// Add validation step at the end
		validationStep := MigrationStep{
			StepType:    MigrationStepValidateData,
			Description: "Validate migrated data integrity",
			Required:    true,
			Reversible:  false,
		}
		plan.Steps = append(plan.Steps, validationStep)
	}

	return plan, nil
}

// Helper methods

func (sm *SchemaManagerImpl) loadSchemas() error {
	files, err := filepath.Glob(filepath.Join(sm.schemaDir, "v*.json"))
	if err != nil {
		return err
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue // Skip files that can't be read
		}

		var schemaVersion SchemaVersion
		if err := json.Unmarshal(data, &schemaVersion); err != nil {
			continue // Skip files that can't be parsed
		}

		// Convert back to Arrow schema
		schema := sm.serializableToSchema(schemaVersion.Schema)
		sm.schemas[schemaVersion.Version] = schema
	}

	return nil
}

func (sm *SchemaManagerImpl) schemaToSerializable(schema *arrow.Schema) *SerializableSchema {
	fields := make([]SerializableField, len(schema.Fields()))
	for i, field := range schema.Fields() {
		fields[i] = SerializableField{
			Name:     field.Name,
			Type:     field.Type.String(),
			Nullable: field.Nullable,
		}
	}
	return &SerializableSchema{Fields: fields}
}

func (sm *SchemaManagerImpl) serializableToSchema(serializable *SerializableSchema) *arrow.Schema {
	fields := make([]arrow.Field, len(serializable.Fields))
	for i, field := range serializable.Fields {
		// This is a simplified conversion - in a real implementation,
		// you'd need to properly parse the type string back to Arrow types
		var dataType arrow.DataType
		switch field.Type {
		case "int64":
			dataType = arrow.PrimitiveTypes.Int64
		case "int32":
			dataType = arrow.PrimitiveTypes.Int32
		case "string":
			dataType = arrow.BinaryTypes.String
		case "bool":
			dataType = arrow.FixedWidthTypes.Boolean
		default:
			dataType = arrow.BinaryTypes.String // Default fallback
		}

		fields[i] = arrow.Field{
			Name:     field.Name,
			Type:     dataType,
			Nullable: field.Nullable,
		}
	}
	return arrow.NewSchema(fields, nil)
}

func (sm *SchemaManagerImpl) getSchemaVersion(schema *arrow.Schema) int {
	// Try to find the version by comparing with known schemas
	for version, knownSchema := range sm.schemas {
		if schema.Equal(knownSchema) {
			return version
		}
	}
	return 0 // Unknown version
}