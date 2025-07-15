package arrow

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/apache/arrow/go/v17/arrow"
)

func TestNewSchemaManager(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()

	t.Run("NewSchemaManager creates valid manager", func(t *testing.T) {
		sm, err := NewSchemaManager(tempDir)
		if err != nil {
			t.Fatalf("NewSchemaManager should not return error: %v", err)
		}

		if sm == nil {
			t.Fatal("NewSchemaManager should not return nil")
		}

		if sm.schemaDir != tempDir {
			t.Errorf("Expected schema directory %s, got %s", tempDir, sm.schemaDir)
		}

		if sm.currentVersion <= 0 {
			t.Errorf("Current version should be positive, got %d", sm.currentVersion)
		}
	})

	t.Run("NewSchemaManager creates directory if not exists", func(t *testing.T) {
		nonExistentDir := filepath.Join(tempDir, "new_schemas")
		sm, err := NewSchemaManager(nonExistentDir)
		if err != nil {
			t.Fatalf("NewSchemaManager should create directory: %v", err)
		}

		if _, err := os.Stat(nonExistentDir); os.IsNotExist(err) {
			t.Error("Schema directory should be created")
		}

		defer sm.GetCurrentSchema() // Just to use sm to avoid unused variable warning
	})

	t.Run("NewSchemaManager with empty directory uses default", func(t *testing.T) {
		sm, err := NewSchemaManager("")
		if err != nil {
			t.Fatalf("NewSchemaManager should handle empty directory: %v", err)
		}

		if sm.schemaDir != "./schemas" {
			t.Errorf("Expected default schema directory './schemas', got %s", sm.schemaDir)
		}

		defer sm.GetCurrentSchema() // Just to use sm to avoid unused variable warning
	})
}

func TestSchemaManagerBasicOperations(t *testing.T) {
	tempDir := t.TempDir()
	sm, err := NewSchemaManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create schema manager: %v", err)
	}

	t.Run("GetCurrentSchema returns valid schema", func(t *testing.T) {
		schema := sm.GetCurrentSchema()
		if schema == nil {
			t.Fatal("GetCurrentSchema should not return nil")
		}

		if !schema.Equal(PlayerSchema) {
			t.Error("GetCurrentSchema should return PlayerSchema")
		}
	})

	t.Run("ValidateSchema with valid schema succeeds", func(t *testing.T) {
		err := sm.ValidateSchema(PlayerSchema)
		if err != nil {
			t.Errorf("ValidateSchema should not return error for valid schema: %v", err)
		}
	})

	t.Run("ValidateSchema with nil schema returns error", func(t *testing.T) {
		err := sm.ValidateSchema(nil)
		if err == nil {
			t.Error("ValidateSchema should return error for nil schema")
		}
	})
}

func TestSchemaRegistration(t *testing.T) {
	tempDir := t.TempDir()
	sm, err := NewSchemaManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create schema manager: %v", err)
	}

	t.Run("RegisterSchema with valid schema succeeds", func(t *testing.T) {
		testSchema := arrow.NewSchema([]arrow.Field{
			{Name: "uid", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
			{Name: "name", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "position", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "age", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "club", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "division", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "nationality", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "nationality_iso", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "transfer_value_amount", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
			{Name: "wage_amount", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
			{Name: "pac", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "sho", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "pas", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "dri", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "def", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "phy", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "overall", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "test_field", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
		}, nil)

		err := sm.RegisterSchema(2, testSchema)
		if err != nil {
			t.Errorf("RegisterSchema should not return error: %v", err)
		}

		// Verify schema was registered
		retrievedSchema, err := sm.GetSchemaByVersion(2)
		if err != nil {
			t.Errorf("GetSchemaByVersion should not return error: %v", err)
		}

		if retrievedSchema == nil {
			t.Fatal("Retrieved schema should not be nil")
		}
	})

	t.Run("RegisterSchema with nil schema returns error", func(t *testing.T) {
		err := sm.RegisterSchema(3, nil)
		if err == nil {
			t.Error("RegisterSchema should return error for nil schema")
		}
	})

	t.Run("GetSchemaByVersion with non-existent version returns error", func(t *testing.T) {
		_, err := sm.GetSchemaByVersion(999)
		if err == nil {
			t.Error("GetSchemaByVersion should return error for non-existent version")
		}
	})
}

func TestSchemaEvolution(t *testing.T) {
	tempDir := t.TempDir()
	sm, err := NewSchemaManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create schema manager: %v", err)
	}

	t.Run("EvolveSchema with nil schemas returns error", func(t *testing.T) {
		_, err := sm.EvolveSchema(nil, PlayerSchema)
		if err == nil {
			t.Error("EvolveSchema should return error for nil old schema")
		}

		_, err = sm.EvolveSchema(PlayerSchema, nil)
		if err == nil {
			t.Error("EvolveSchema should return error for nil new schema")
		}
	})

	t.Run("EvolveSchema with identical schemas", func(t *testing.T) {
		evolution, err := sm.EvolveSchema(PlayerSchema, PlayerSchema)
		if err != nil {
			t.Errorf("EvolveSchema should not return error for identical schemas: %v", err)
		}

		if !evolution.Compatible {
			t.Error("Identical schemas should be compatible")
		}

		if len(evolution.AddedFields) != 0 {
			t.Errorf("Identical schemas should have no added fields, got %d", len(evolution.AddedFields))
		}

		if len(evolution.RemovedFields) != 0 {
			t.Errorf("Identical schemas should have no removed fields, got %d", len(evolution.RemovedFields))
		}

		if len(evolution.ModifiedFields) != 0 {
			t.Errorf("Identical schemas should have no modified fields, got %d", len(evolution.ModifiedFields))
		}
	})

	t.Run("EvolveSchema with added field", func(t *testing.T) {
		oldSchema := arrow.NewSchema([]arrow.Field{
			{Name: "uid", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
			{Name: "name", Type: arrow.BinaryTypes.String, Nullable: false},
		}, nil)

		newSchema := arrow.NewSchema([]arrow.Field{
			{Name: "uid", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
			{Name: "name", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "new_field", Type: arrow.PrimitiveTypes.Int32, Nullable: true},
		}, nil)

		evolution, err := sm.EvolveSchema(oldSchema, newSchema)
		if err != nil {
			t.Errorf("EvolveSchema should not return error: %v", err)
		}

		if !evolution.Compatible {
			t.Error("Adding nullable field should be compatible")
		}

		if len(evolution.AddedFields) != 1 {
			t.Errorf("Expected 1 added field, got %d", len(evolution.AddedFields))
		}

		if evolution.AddedFields[0].Name != "new_field" {
			t.Errorf("Expected added field 'new_field', got '%s'", evolution.AddedFields[0].Name)
		}
	})

	t.Run("EvolveSchema with removed field", func(t *testing.T) {
		oldSchema := arrow.NewSchema([]arrow.Field{
			{Name: "uid", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
			{Name: "name", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "removed_field", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
		}, nil)

		newSchema := arrow.NewSchema([]arrow.Field{
			{Name: "uid", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
			{Name: "name", Type: arrow.BinaryTypes.String, Nullable: false},
		}, nil)

		evolution, err := sm.EvolveSchema(oldSchema, newSchema)
		if err != nil {
			t.Errorf("EvolveSchema should not return error: %v", err)
		}

		if evolution.Compatible {
			t.Error("Removing field should not be compatible")
		}

		if len(evolution.RemovedFields) != 1 {
			t.Errorf("Expected 1 removed field, got %d", len(evolution.RemovedFields))
		}

		if evolution.RemovedFields[0].Name != "removed_field" {
			t.Errorf("Expected removed field 'removed_field', got '%s'", evolution.RemovedFields[0].Name)
		}
	})

	t.Run("EvolveSchema with compatible type change", func(t *testing.T) {
		oldSchema := arrow.NewSchema([]arrow.Field{
			{Name: "test_field", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
		}, nil)

		newSchema := arrow.NewSchema([]arrow.Field{
			{Name: "test_field", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
		}, nil)

		evolution, err := sm.EvolveSchema(oldSchema, newSchema)
		if err != nil {
			t.Errorf("EvolveSchema should not return error: %v", err)
		}

		if !evolution.Compatible {
			t.Error("Compatible type change should be compatible")
		}

		if len(evolution.ModifiedFields) != 1 {
			t.Errorf("Expected 1 modified field, got %d", len(evolution.ModifiedFields))
		}

		modification := evolution.ModifiedFields[0]
		if !modification.Compatible {
			t.Error("INT32 to INT64 should be compatible")
		}
	})

	t.Run("EvolveSchema with incompatible type change", func(t *testing.T) {
		oldSchema := arrow.NewSchema([]arrow.Field{
			{Name: "test_field", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
		}, nil)

		newSchema := arrow.NewSchema([]arrow.Field{
			{Name: "test_field", Type: arrow.BinaryTypes.String, Nullable: false},
		}, nil)

		evolution, err := sm.EvolveSchema(oldSchema, newSchema)
		if err != nil {
			t.Errorf("EvolveSchema should not return error: %v", err)
		}

		if evolution.Compatible {
			t.Error("Incompatible type change should not be compatible")
		}

		if len(evolution.ModifiedFields) != 1 {
			t.Errorf("Expected 1 modified field, got %d", len(evolution.ModifiedFields))
		}

		modification := evolution.ModifiedFields[0]
		if modification.Compatible {
			t.Error("INT64 to STRING should not be compatible")
		}
	})
}

func TestSchemaHistory(t *testing.T) {
	tempDir := t.TempDir()
	sm, err := NewSchemaManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create schema manager: %v", err)
	}

	t.Run("GetSchemaHistory returns current schema", func(t *testing.T) {
		history, err := sm.GetSchemaHistory()
		if err != nil {
			t.Errorf("GetSchemaHistory should not return error: %v", err)
		}

		if len(history) == 0 {
			t.Error("History should contain at least the current schema")
		}

		// Find current version in history
		found := false
		for _, version := range history {
			if version.Version == sm.currentVersion {
				found = true
				break
			}
		}

		if !found {
			t.Error("Current schema version should be in history")
		}
	})

	t.Run("GetSchemaHistory after registering multiple schemas", func(t *testing.T) {
		// Register additional schemas with required fields
		schema2 := arrow.NewSchema([]arrow.Field{
			{Name: "uid", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
			{Name: "name", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "position", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "age", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "club", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "division", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "nationality", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "nationality_iso", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "transfer_value_amount", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
			{Name: "wage_amount", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
			{Name: "pac", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "sho", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "pas", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "dri", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "def", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "phy", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "overall", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
		}, nil)

		schema3 := arrow.NewSchema([]arrow.Field{
			{Name: "uid", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
			{Name: "name", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "position", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "age", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "club", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "division", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "nationality", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "nationality_iso", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "transfer_value_amount", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
			{Name: "wage_amount", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
			{Name: "pac", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "sho", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "pas", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "dri", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "def", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "phy", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "overall", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "new_field", Type: arrow.BinaryTypes.String, Nullable: true},
		}, nil)

		err := sm.RegisterSchema(2, schema2)
		if err != nil {
			t.Errorf("Failed to register schema 2: %v", err)
		}

		err = sm.RegisterSchema(3, schema3)
		if err != nil {
			t.Errorf("Failed to register schema 3: %v", err)
		}

		history, err := sm.GetSchemaHistory()
		if err != nil {
			t.Errorf("GetSchemaHistory should not return error: %v", err)
		}

		if len(history) < 3 {
			t.Errorf("History should contain at least 3 schemas, got %d", len(history))
		}

		// Verify history is sorted by version
		for i := 1; i < len(history); i++ {
			if history[i-1].Version >= history[i].Version {
				t.Error("History should be sorted by version")
				break
			}
		}
	})
}

func TestMigrationPlanning(t *testing.T) {
	tempDir := t.TempDir()
	sm, err := NewSchemaManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create schema manager: %v", err)
	}

	// Register test schemas with all required fields
	schema1 := arrow.NewSchema([]arrow.Field{
		{Name: "uid", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
		{Name: "name", Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: "position", Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: "age", Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: "club", Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: "division", Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: "nationality", Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: "nationality_iso", Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: "transfer_value_amount", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
		{Name: "wage_amount", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
		{Name: "pac", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
		{Name: "sho", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
		{Name: "pas", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
		{Name: "dri", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
		{Name: "def", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
		{Name: "phy", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
		{Name: "overall", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
	}, nil)

	schema2 := arrow.NewSchema([]arrow.Field{
		{Name: "uid", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
		{Name: "name", Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: "position", Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: "age", Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: "club", Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: "division", Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: "nationality", Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: "nationality_iso", Type: arrow.BinaryTypes.String, Nullable: false},
		{Name: "transfer_value_amount", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
		{Name: "wage_amount", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
		{Name: "pac", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
		{Name: "sho", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
		{Name: "pas", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
		{Name: "dri", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
		{Name: "def", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
		{Name: "phy", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
		{Name: "overall", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
		{Name: "extra_field", Type: arrow.PrimitiveTypes.Int32, Nullable: true},
	}, nil)

	err = sm.RegisterSchema(10, schema1)
	if err != nil {
		t.Fatalf("Failed to register schema 1: %v", err)
	}

	err = sm.RegisterSchema(11, schema2)
	if err != nil {
		t.Fatalf("Failed to register schema 2: %v", err)
	}

	t.Run("PlanMigration with valid versions", func(t *testing.T) {
		plan, err := sm.PlanMigration(10, 11)
		if err != nil {
			t.Errorf("PlanMigration should not return error: %v", err)
		}

		if plan == nil {
			t.Fatal("Migration plan should not be nil")
		}

		if plan.FromVersion != 10 {
			t.Errorf("Expected from version 10, got %d", plan.FromVersion)
		}

		if plan.ToVersion != 11 {
			t.Errorf("Expected to version 11, got %d", plan.ToVersion)
		}

		if len(plan.Steps) == 0 {
			t.Error("Migration plan should have steps")
		}

		// Should have backup, add field, and validation steps
		expectedStepTypes := []MigrationStepType{
			MigrationStepBackupData,
			MigrationStepAddField,
			MigrationStepValidateData,
		}

		if len(plan.Steps) != len(expectedStepTypes) {
			t.Errorf("Expected %d steps, got %d", len(expectedStepTypes), len(plan.Steps))
		}

		for i, expectedType := range expectedStepTypes {
			if i < len(plan.Steps) && plan.Steps[i].StepType != expectedType {
				t.Errorf("Expected step %d to be %s, got %s", i, expectedType.String(), plan.Steps[i].StepType.String())
			}
		}
	})

	t.Run("PlanMigration with non-existent source version", func(t *testing.T) {
		_, err := sm.PlanMigration(999, 11)
		if err == nil {
			t.Error("PlanMigration should return error for non-existent source version")
		}
	})

	t.Run("PlanMigration with non-existent target version", func(t *testing.T) {
		_, err := sm.PlanMigration(10, 999)
		if err == nil {
			t.Error("PlanMigration should return error for non-existent target version")
		}
	})
}

func TestMigrationStepType(t *testing.T) {
	testCases := []struct {
		stepType MigrationStepType
		expected string
	}{
		{MigrationStepAddField, "ADD_FIELD"},
		{MigrationStepRemoveField, "REMOVE_FIELD"},
		{MigrationStepModifyField, "MODIFY_FIELD"},
		{MigrationStepValidateData, "VALIDATE_DATA"},
		{MigrationStepBackupData, "BACKUP_DATA"},
		{MigrationStepTransformData, "TRANSFORM_DATA"},
		{MigrationStepType(999), "UNKNOWN"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := tc.stepType.String()
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestSchemaEvolutionFields(t *testing.T) {
	tempDir := t.TempDir()
	sm, err := NewSchemaManager(tempDir)
	if err != nil {
		t.Fatalf("Failed to create schema manager: %v", err)
	}

	t.Run("SchemaEvolution has correct timestamps", func(t *testing.T) {
		oldSchema := arrow.NewSchema([]arrow.Field{
			{Name: "test_field", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
		}, nil)

		newSchema := arrow.NewSchema([]arrow.Field{
			{Name: "test_field", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "new_field", Type: arrow.BinaryTypes.String, Nullable: true},
		}, nil)

		before := time.Now()
		evolution, err := sm.EvolveSchema(oldSchema, newSchema)
		after := time.Now()

		if err != nil {
			t.Errorf("EvolveSchema should not return error: %v", err)
		}

		if evolution.CreatedAt.Before(before) || evolution.CreatedAt.After(after) {
			t.Error("Evolution CreatedAt should be set to current time")
		}
	})

	t.Run("MigrationPlan has correct structure", func(t *testing.T) {
		schema1 := arrow.NewSchema([]arrow.Field{
			{Name: "uid", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
			{Name: "name", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "position", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "age", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "club", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "division", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "nationality", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "nationality_iso", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "transfer_value_amount", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
			{Name: "wage_amount", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
			{Name: "pac", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "sho", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "pas", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "dri", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "def", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "phy", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "overall", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
		}, nil)

		schema2 := arrow.NewSchema([]arrow.Field{
			{Name: "uid", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
			{Name: "name", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "position", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "age", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "club", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "division", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "nationality", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "nationality_iso", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "transfer_value_amount", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
			{Name: "wage_amount", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
			{Name: "pac", Type: arrow.PrimitiveTypes.Int64, Nullable: false}, // Changed to INT64 for type promotion test
			{Name: "sho", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "pas", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "dri", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "def", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "phy", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
			{Name: "overall", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
		}, nil)

		err := sm.RegisterSchema(20, schema1)
		if err != nil {
			t.Fatalf("Failed to register schema: %v", err)
		}

		err = sm.RegisterSchema(21, schema2)
		if err != nil {
			t.Fatalf("Failed to register schema: %v", err)
		}

		before := time.Now()
		plan, err := sm.PlanMigration(20, 21)
		after := time.Now()

		if err != nil {
			t.Errorf("PlanMigration should not return error: %v", err)
		}

		if plan.CreatedAt.Before(before) || plan.CreatedAt.After(after) {
			t.Error("Plan CreatedAt should be set to current time")
		}

		if plan.EstimatedDuration <= 0 {
			t.Error("Plan should have positive estimated duration")
		}
	})
}