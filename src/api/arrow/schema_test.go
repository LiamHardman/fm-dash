package arrow

import (
	"testing"

	"github.com/apache/arrow/go/v17/arrow"
)

func TestPlayerSchema(t *testing.T) {
	t.Run("PlayerSchema is not nil", func(t *testing.T) {
		if PlayerSchema == nil {
			t.Fatal("PlayerSchema should not be nil")
		}
	})

	t.Run("PlayerSchema has expected number of fields", func(t *testing.T) {
		expectedFieldCount := 39 // Based on the actual schema definition
		if len(PlayerSchema.Fields()) != expectedFieldCount {
			t.Errorf("Expected %d fields, got %d", expectedFieldCount, len(PlayerSchema.Fields()))
		}
	})

	t.Run("PlayerSchema has required fields", func(t *testing.T) {
		requiredFields := map[string]arrow.Type{
			"uid":                   arrow.INT64,
			"name":                  arrow.STRING,
			"position":              arrow.STRING,
			"age":                   arrow.STRING,
			"club":                  arrow.STRING,
			"division":              arrow.STRING,
			"nationality":           arrow.STRING,
			"nationality_iso":       arrow.STRING,
			"transfer_value_amount": arrow.INT64,
			"wage_amount":           arrow.INT64,
			"pac":                   arrow.INT32,
			"sho":                   arrow.INT32,
			"pas":                   arrow.INT32,
			"dri":                   arrow.INT32,
			"def":                   arrow.INT32,
			"phy":                   arrow.INT32,
			"overall":               arrow.INT32,
		}

		schemaFields := make(map[string]arrow.Field)
		for i, field := range PlayerSchema.Fields() {
			schemaFields[field.Name] = PlayerSchema.Field(i)
		}

		for fieldName, expectedType := range requiredFields {
			field, exists := schemaFields[fieldName]
			if !exists {
				t.Errorf("Required field '%s' is missing from schema", fieldName)
				continue
			}

			if field.Type.ID() != expectedType {
				t.Errorf("Field '%s' has type %s, expected %s", fieldName, field.Type.String(), expectedType.String())
			}
		}
	})

	t.Run("PlayerSchema nullable fields are correctly set", func(t *testing.T) {
		nonNullableFields := []string{
			"uid", "name", "position", "age", "club", "division",
			"nationality", "nationality_iso", "transfer_value_amount",
			"wage_amount", "pac", "sho", "pas", "dri", "def", "phy", "overall",
		}

		nullableFields := []string{
			"gk", "div", "han", "ref", "kic", "spd", "pos",
			"personality", "media_handling", "attribute_masked",
		}

		schemaFields := make(map[string]arrow.Field)
		for i, field := range PlayerSchema.Fields() {
			schemaFields[field.Name] = PlayerSchema.Field(i)
		}

		for _, fieldName := range nonNullableFields {
			if field, exists := schemaFields[fieldName]; exists {
				if field.Nullable {
					t.Errorf("Field '%s' should not be nullable", fieldName)
				}
			}
		}

		for _, fieldName := range nullableFields {
			if field, exists := schemaFields[fieldName]; exists {
				if !field.Nullable {
					t.Errorf("Field '%s' should be nullable", fieldName)
				}
			}
		}
	})
}

func TestSchemaValidator(t *testing.T) {
	validator := NewSchemaValidator()

	t.Run("NewSchemaValidator creates valid validator", func(t *testing.T) {
		if validator == nil {
			t.Fatal("NewSchemaValidator should not return nil")
		}
		if validator.pool == nil {
			t.Fatal("SchemaValidator should have a memory pool")
		}
	})

	t.Run("ValidateSchema with nil schema returns error", func(t *testing.T) {
		err := validator.ValidateSchema(nil)
		if err == nil {
			t.Error("ValidateSchema should return error for nil schema")
		}
	})

	t.Run("ValidateSchema with valid PlayerSchema succeeds", func(t *testing.T) {
		err := validator.ValidateSchema(PlayerSchema)
		if err != nil {
			t.Errorf("ValidateSchema should not return error for valid PlayerSchema: %v", err)
		}
	})

	t.Run("ValidateSchema with missing required field returns error", func(t *testing.T) {
		// Create schema missing a required field
		incompleteSchema := arrow.NewSchema([]arrow.Field{
			{Name: "uid", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
			{Name: "name", Type: arrow.BinaryTypes.String, Nullable: false},
			// Missing other required fields
		}, nil)

		err := validator.ValidateSchema(incompleteSchema)
		if err == nil {
			t.Error("ValidateSchema should return error for schema missing required fields")
		}
	})
}

func TestSchemaCompatibility(t *testing.T) {
	validator := NewSchemaValidator()

	t.Run("ValidateSchemaCompatibility with nil schemas returns error", func(t *testing.T) {
		err := validator.ValidateSchemaCompatibility(nil, PlayerSchema)
		if err == nil {
			t.Error("ValidateSchemaCompatibility should return error for nil old schema")
		}

		err = validator.ValidateSchemaCompatibility(PlayerSchema, nil)
		if err == nil {
			t.Error("ValidateSchemaCompatibility should return error for nil new schema")
		}
	})

	t.Run("ValidateSchemaCompatibility with identical schemas succeeds", func(t *testing.T) {
		err := validator.ValidateSchemaCompatibility(PlayerSchema, PlayerSchema)
		if err != nil {
			t.Errorf("ValidateSchemaCompatibility should not return error for identical schemas: %v", err)
		}
	})

	t.Run("ValidateSchemaCompatibility with compatible type promotion succeeds", func(t *testing.T) {
		oldSchema := arrow.NewSchema([]arrow.Field{
			{Name: "test_field", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
		}, nil)

		newSchema := arrow.NewSchema([]arrow.Field{
			{Name: "test_field", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
		}, nil)

		err := validator.ValidateSchemaCompatibility(oldSchema, newSchema)
		if err != nil {
			t.Errorf("ValidateSchemaCompatibility should allow INT32 to INT64 promotion: %v", err)
		}
	})

	t.Run("ValidateSchemaCompatibility with incompatible type change returns error", func(t *testing.T) {
		oldSchema := arrow.NewSchema([]arrow.Field{
			{Name: "test_field", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
		}, nil)

		newSchema := arrow.NewSchema([]arrow.Field{
			{Name: "test_field", Type: arrow.BinaryTypes.String, Nullable: false},
		}, nil)

		err := validator.ValidateSchemaCompatibility(oldSchema, newSchema)
		if err == nil {
			t.Error("ValidateSchemaCompatibility should return error for incompatible type change")
		}
	})

	t.Run("ValidateSchemaCompatibility with nullable to non-nullable change returns error", func(t *testing.T) {
		oldSchema := arrow.NewSchema([]arrow.Field{
			{Name: "test_field", Type: arrow.PrimitiveTypes.Int32, Nullable: true},
		}, nil)

		newSchema := arrow.NewSchema([]arrow.Field{
			{Name: "test_field", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
		}, nil)

		err := validator.ValidateSchemaCompatibility(oldSchema, newSchema)
		if err == nil {
			t.Error("ValidateSchemaCompatibility should return error for nullable to non-nullable change")
		}
	})

	t.Run("ValidateSchemaCompatibility with non-nullable to nullable change succeeds", func(t *testing.T) {
		oldSchema := arrow.NewSchema([]arrow.Field{
			{Name: "test_field", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
		}, nil)

		newSchema := arrow.NewSchema([]arrow.Field{
			{Name: "test_field", Type: arrow.PrimitiveTypes.Int32, Nullable: true},
		}, nil)

		err := validator.ValidateSchemaCompatibility(oldSchema, newSchema)
		if err != nil {
			t.Errorf("ValidateSchemaCompatibility should allow non-nullable to nullable change: %v", err)
		}
	})
}

func TestTypeCompatibility(t *testing.T) {
	validator := NewSchemaValidator()

	testCases := []struct {
		name       string
		oldType    arrow.DataType
		newType    arrow.DataType
		compatible bool
	}{
		{
			name:       "INT32 to INT64 is compatible",
			oldType:    arrow.PrimitiveTypes.Int32,
			newType:    arrow.PrimitiveTypes.Int64,
			compatible: true,
		},
		{
			name:       "FLOAT32 to FLOAT64 is compatible",
			oldType:    arrow.PrimitiveTypes.Float32,
			newType:    arrow.PrimitiveTypes.Float64,
			compatible: true,
		},
		{
			name:       "STRING to LARGE_STRING is compatible",
			oldType:    arrow.BinaryTypes.String,
			newType:    arrow.BinaryTypes.LargeString,
			compatible: true,
		},
		{
			name:       "INT64 to INT32 is not compatible",
			oldType:    arrow.PrimitiveTypes.Int64,
			newType:    arrow.PrimitiveTypes.Int32,
			compatible: false,
		},
		{
			name:       "STRING to INT32 is not compatible",
			oldType:    arrow.BinaryTypes.String,
			newType:    arrow.PrimitiveTypes.Int32,
			compatible: false,
		},
		{
			name:       "Identical types are compatible",
			oldType:    arrow.PrimitiveTypes.Int32,
			newType:    arrow.PrimitiveTypes.Int32,
			compatible: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := validator.areTypesCompatible(tc.oldType, tc.newType)
			if result != tc.compatible {
				t.Errorf("Expected compatibility %v, got %v for %s -> %s",
					tc.compatible, result, tc.oldType.String(), tc.newType.String())
			}
		})
	}
}

func TestCreateEmptyTable(t *testing.T) {
	t.Run("CreateEmptyTable returns valid table", func(t *testing.T) {
		table := CreateEmptyTable()
		defer table.Release()

		if table.NumRows() != 0 {
			t.Errorf("Empty table should have 0 rows, got %d", table.NumRows())
		}

		if table.NumCols() != int64(len(PlayerSchema.Fields())) {
			t.Errorf("Table should have %d columns, got %d", len(PlayerSchema.Fields()), table.NumCols())
		}

		if !table.Schema().Equal(PlayerSchema) {
			t.Error("Table schema should match PlayerSchema")
		}
	})
}

func TestSchemaUtilities(t *testing.T) {
	t.Run("GetSchemaVersion returns valid version", func(t *testing.T) {
		version := GetSchemaVersion()
		if version <= 0 {
			t.Errorf("Schema version should be positive, got %d", version)
		}
	})

	t.Run("GetSchemaFingerprint with nil schema returns empty string", func(t *testing.T) {
		fingerprint := GetSchemaFingerprint(nil)
		if fingerprint != "" {
			t.Errorf("Fingerprint for nil schema should be empty, got '%s'", fingerprint)
		}
	})

	t.Run("GetSchemaFingerprint returns non-empty string for valid schema", func(t *testing.T) {
		fingerprint := GetSchemaFingerprint(PlayerSchema)
		if fingerprint == "" {
			t.Error("Fingerprint for valid schema should not be empty")
		}
	})

	t.Run("GetSchemaFingerprint is consistent", func(t *testing.T) {
		fingerprint1 := GetSchemaFingerprint(PlayerSchema)
		fingerprint2 := GetSchemaFingerprint(PlayerSchema)
		if fingerprint1 != fingerprint2 {
			t.Error("Schema fingerprint should be consistent")
		}
	})
}

func TestCompareSchemas(t *testing.T) {
	t.Run("CompareSchemas with nil schemas", func(t *testing.T) {
		comparison := CompareSchemas(nil, PlayerSchema)
		if comparison.Compatible {
			t.Error("Comparison with nil schema should not be compatible")
		}
		if len(comparison.Changes) == 0 {
			t.Error("Comparison with nil schema should have changes")
		}
	})

	t.Run("CompareSchemas with identical schemas", func(t *testing.T) {
		comparison := CompareSchemas(PlayerSchema, PlayerSchema)
		if !comparison.Compatible {
			t.Error("Identical schemas should be compatible")
		}
		if len(comparison.Changes) != 0 {
			t.Errorf("Identical schemas should have no changes, got %d", len(comparison.Changes))
		}
	})

	t.Run("CompareSchemas with added field", func(t *testing.T) {
		oldSchema := arrow.NewSchema([]arrow.Field{
			{Name: "uid", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
			{Name: "name", Type: arrow.BinaryTypes.String, Nullable: false},
		}, nil)

		newSchema := arrow.NewSchema([]arrow.Field{
			{Name: "uid", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
			{Name: "name", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "new_field", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
		}, nil)

		comparison := CompareSchemas(oldSchema, newSchema)
		if !comparison.Compatible {
			t.Error("Adding field should be compatible")
		}

		foundAddedChange := false
		for _, change := range comparison.Changes {
			if change.Type == SchemaChangeAdded && change.FieldName == "new_field" {
				foundAddedChange = true
				break
			}
		}
		if !foundAddedChange {
			t.Error("Should detect added field")
		}
	})

	t.Run("CompareSchemas with removed field", func(t *testing.T) {
		oldSchema := arrow.NewSchema([]arrow.Field{
			{Name: "uid", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
			{Name: "name", Type: arrow.BinaryTypes.String, Nullable: false},
			{Name: "removed_field", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
		}, nil)

		newSchema := arrow.NewSchema([]arrow.Field{
			{Name: "uid", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
			{Name: "name", Type: arrow.BinaryTypes.String, Nullable: false},
		}, nil)

		comparison := CompareSchemas(oldSchema, newSchema)
		if comparison.Compatible {
			t.Error("Removing field should not be compatible")
		}

		foundRemovedChange := false
		for _, change := range comparison.Changes {
			if change.Type == SchemaChangeRemoved && change.FieldName == "removed_field" {
				foundRemovedChange = true
				break
			}
		}
		if !foundRemovedChange {
			t.Error("Should detect removed field")
		}
	})

	t.Run("CompareSchemas with compatible type change", func(t *testing.T) {
		oldSchema := arrow.NewSchema([]arrow.Field{
			{Name: "test_field", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
		}, nil)

		newSchema := arrow.NewSchema([]arrow.Field{
			{Name: "test_field", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
		}, nil)

		comparison := CompareSchemas(oldSchema, newSchema)
		if !comparison.Compatible {
			t.Error("Compatible type change should be compatible")
		}

		foundModifiedChange := false
		for _, change := range comparison.Changes {
			if change.Type == SchemaChangeModified && change.FieldName == "test_field" {
				foundModifiedChange = true
				break
			}
		}
		if !foundModifiedChange {
			t.Error("Should detect modified field")
		}
	})
}

func TestSchemaChangeType(t *testing.T) {
	testCases := []struct {
		changeType SchemaChangeType
		expected   string
	}{
		{SchemaChangeAdded, "ADDED"},
		{SchemaChangeRemoved, "REMOVED"},
		{SchemaChangeModified, "MODIFIED"},
		{SchemaChangeInvalid, "INVALID"},
		{SchemaChangeType(999), "UNKNOWN"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := tc.changeType.String()
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}