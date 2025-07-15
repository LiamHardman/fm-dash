package main

import (
	"testing"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/memory"
)

func TestNewPlayerArrowSchema(t *testing.T) {
	schema := NewPlayerArrowSchema()
	
	if schema == nil {
		t.Fatal("Expected schema to be created, got nil")
	}
	
	if schema.GetVersion() != CurrentSchemaVersion {
		t.Errorf("Expected schema version %s, got %s", CurrentSchemaVersion, schema.GetVersion())
	}
	
	if schema.GetSchema() == nil {
		t.Fatal("Expected Arrow schema to be created, got nil")
	}
}

func TestPlayerSchemaFields(t *testing.T) {
	schema := NewPlayerArrowSchema()
	arrowSchema := schema.GetSchema()
	
	// Test that we have the expected number of fields
	expectedMinFields := 25 // Minimum expected fields
	if len(arrowSchema.Fields()) < expectedMinFields {
		t.Errorf("Expected at least %d fields, got %d", expectedMinFields, len(arrowSchema.Fields()))
	}
	
	// Test required fields exist with correct types
	testCases := []struct {
		fieldName    string
		expectedType arrow.DataType
		nullable     bool
	}{
		{"uid", arrow.PrimitiveTypes.Int64, false},
		{"overall", arrow.PrimitiveTypes.Int32, false},
		{"age_numeric", arrow.PrimitiveTypes.Int32, false},
		{"transfer_value_amount", arrow.PrimitiveTypes.Int64, false},
		{"wage_amount", arrow.PrimitiveTypes.Int64, false},
		{"pac", arrow.PrimitiveTypes.Int16, false},
		{"sho", arrow.PrimitiveTypes.Int16, false},
		{"pas", arrow.PrimitiveTypes.Int16, false},
		{"dri", arrow.PrimitiveTypes.Int16, false},
		{"def", arrow.PrimitiveTypes.Int16, false},
		{"phy", arrow.PrimitiveTypes.Int16, false},
		{"gk", arrow.PrimitiveTypes.Int16, true},
		{"div", arrow.PrimitiveTypes.Int16, true},
		{"han", arrow.PrimitiveTypes.Int16, true},
		{"ref", arrow.PrimitiveTypes.Int16, true},
		{"kic", arrow.PrimitiveTypes.Int16, true},
		{"spd", arrow.PrimitiveTypes.Int16, true},
		{"pos", arrow.PrimitiveTypes.Int16, true},
		{"attribute_masked", arrow.FixedWidthTypes.Boolean, false},
		{"personality", arrow.BinaryTypes.String, true},
		{"media_handling", arrow.BinaryTypes.String, true},
		{"best_role_overall", arrow.BinaryTypes.String, true},
	}
	
	for _, tc := range testCases {
		field, found := schema.GetFieldByName(tc.fieldName)
		if !found {
			t.Errorf("Expected field '%s' to exist", tc.fieldName)
			continue
		}
		
		if !arrow.TypeEqual(field.Type, tc.expectedType) {
			t.Errorf("Field '%s': expected type %s, got %s", tc.fieldName, tc.expectedType, field.Type)
		}
		
		if field.Nullable != tc.nullable {
			t.Errorf("Field '%s': expected nullable %t, got %t", tc.fieldName, tc.nullable, field.Nullable)
		}
	}
}

func TestDictionaryEncodedFields(t *testing.T) {
	schema := NewPlayerArrowSchema()
	
	// Test dictionary-encoded string fields
	dictionaryFields := []struct {
		fieldName         string
		expectedIndexType arrow.DataType
		expectedValueType arrow.DataType
	}{
		{"name", arrow.PrimitiveTypes.Uint16, arrow.BinaryTypes.String},
		{"position", arrow.PrimitiveTypes.Uint8, arrow.BinaryTypes.String},
		{"club", arrow.PrimitiveTypes.Uint16, arrow.BinaryTypes.String},
		{"division", arrow.PrimitiveTypes.Uint8, arrow.BinaryTypes.String},
		{"nationality", arrow.PrimitiveTypes.Uint8, arrow.BinaryTypes.String},
		{"nationality_iso", arrow.PrimitiveTypes.Uint8, arrow.BinaryTypes.String},
		{"nationality_fifa_code", arrow.PrimitiveTypes.Uint8, arrow.BinaryTypes.String},
	}
	
	for _, tc := range dictionaryFields {
		field, found := schema.GetFieldByName(tc.fieldName)
		if !found {
			t.Errorf("Expected dictionary field '%s' to exist", tc.fieldName)
			continue
		}
		
		dictType, ok := field.Type.(*arrow.DictionaryType)
		if !ok {
			t.Errorf("Field '%s': expected dictionary type, got %T", tc.fieldName, field.Type)
			continue
		}
		
		if !arrow.TypeEqual(dictType.IndexType, tc.expectedIndexType) {
			t.Errorf("Field '%s': expected index type %s, got %s", tc.fieldName, tc.expectedIndexType, dictType.IndexType)
		}
		
		if !arrow.TypeEqual(dictType.ValueType, tc.expectedValueType) {
			t.Errorf("Field '%s': expected value type %s, got %s", tc.fieldName, tc.expectedValueType, dictType.ValueType)
		}
	}
}

func TestListFields(t *testing.T) {
	schema := NewPlayerArrowSchema()
	
	// Test list fields
	listFields := []string{"parsed_positions", "short_positions", "position_groups"}
	
	for _, fieldName := range listFields {
		field, found := schema.GetFieldByName(fieldName)
		if !found {
			t.Errorf("Expected list field '%s' to exist", fieldName)
			continue
		}
		
		listType, ok := field.Type.(*arrow.ListType)
		if !ok {
			t.Errorf("Field '%s': expected list type, got %T", fieldName, field.Type)
			continue
		}
		
		// Check that list element is dictionary-encoded string
		dictType, ok := listType.Elem().(*arrow.DictionaryType)
		if !ok {
			t.Errorf("Field '%s': expected list of dictionary type, got %T", fieldName, listType.Elem())
			continue
		}
		
		if !arrow.TypeEqual(dictType.IndexType, arrow.PrimitiveTypes.Uint8) {
			t.Errorf("Field '%s': expected list element index type Uint8, got %s", fieldName, dictType.IndexType)
		}
		
		if !arrow.TypeEqual(dictType.ValueType, arrow.BinaryTypes.String) {
			t.Errorf("Field '%s': expected list element value type String, got %s", fieldName, dictType.ValueType)
		}
	}
}

func TestStructFields(t *testing.T) {
	schema := NewPlayerArrowSchema()
	
	// Test performance_stats struct field
	field, found := schema.GetFieldByName("performance_stats")
	if !found {
		t.Fatal("Expected 'performance_stats' field to exist")
	}
	
	structType, ok := field.Type.(*arrow.StructType)
	if !ok {
		t.Fatalf("Expected 'performance_stats' to be struct type, got %T", field.Type)
	}
	
	// Check struct fields
	expectedStructFields := []string{"goals", "assists", "pass_completion", "shots_per_game", "tackles_per_game"}
	if len(structType.Fields()) != len(expectedStructFields) {
		t.Errorf("Expected %d struct fields, got %d", len(expectedStructFields), len(structType.Fields()))
	}
	
	for _, expectedField := range expectedStructFields {
		found := false
		for _, structField := range structType.Fields() {
			if structField.Name == expectedField {
				found = true
				if !arrow.TypeEqual(structField.Type, arrow.PrimitiveTypes.Float64) {
					t.Errorf("Struct field '%s': expected Float64 type, got %s", expectedField, structField.Type)
				}
				if !structField.Nullable {
					t.Errorf("Struct field '%s': expected to be nullable", expectedField)
				}
				break
			}
		}
		if !found {
			t.Errorf("Expected struct field '%s' not found", expectedField)
		}
	}
	
	// Test role_specific_overalls list of struct field
	field, found = schema.GetFieldByName("role_specific_overalls")
	if !found {
		t.Fatal("Expected 'role_specific_overalls' field to exist")
	}
	
	listType, ok := field.Type.(*arrow.ListType)
	if !ok {
		t.Fatalf("Expected 'role_specific_overalls' to be list type, got %T", field.Type)
	}
	
	structType, ok = listType.Elem().(*arrow.StructType)
	if !ok {
		t.Fatalf("Expected 'role_specific_overalls' list element to be struct type, got %T", listType.Elem())
	}
	
	// Check role struct fields
	expectedRoleFields := map[string]arrow.DataType{
		"role_name": arrow.BinaryTypes.String,
		"score":     arrow.PrimitiveTypes.Int32,
	}
	
	if len(structType.Fields()) != len(expectedRoleFields) {
		t.Errorf("Expected %d role struct fields, got %d", len(expectedRoleFields), len(structType.Fields()))
	}
	
	for _, structField := range structType.Fields() {
		expectedType, exists := expectedRoleFields[structField.Name]
		if !exists {
			t.Errorf("Unexpected role struct field '%s'", structField.Name)
			continue
		}
		
		if !arrow.TypeEqual(structField.Type, expectedType) {
			t.Errorf("Role struct field '%s': expected type %s, got %s", structField.Name, expectedType, structField.Type)
		}
	}
}

func TestSchemaMetadata(t *testing.T) {
	schema := NewPlayerArrowSchema()
	arrowSchema := schema.GetSchema()
	
	metadata := arrowSchema.Metadata()
	if metadata.Len() == 0 {
		t.Fatal("Expected schema to have metadata")
	}
	
	// Check required metadata keys
	requiredKeys := []string{"schema_version", "created_at", "format_type"}
	for _, key := range requiredKeys {
		found := false
		for i := 0; i < metadata.Len(); i++ {
			if metadata.Keys()[i] == key {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected metadata key '%s' to exist", key)
		}
	}
	
	// Check schema version in metadata
	versionFound := false
	for i := 0; i < metadata.Len(); i++ {
		if metadata.Keys()[i] == "schema_version" {
			if metadata.Values()[i] != string(CurrentSchemaVersion) {
				t.Errorf("Expected metadata schema_version %s, got %s", CurrentSchemaVersion, metadata.Values()[i])
			}
			versionFound = true
			break
		}
	}
	if !versionFound {
		t.Error("Expected to find schema_version in metadata")
	}
	
	// Check format type
	formatFound := false
	for i := 0; i < metadata.Len(); i++ {
		if metadata.Keys()[i] == "format_type" {
			if metadata.Values()[i] != "player_data" {
				t.Errorf("Expected metadata format_type 'player_data', got %s", metadata.Values()[i])
			}
			formatFound = true
			break
		}
	}
	if !formatFound {
		t.Error("Expected to find format_type in metadata")
	}
}

func TestValidateSchema(t *testing.T) {
	schema := NewPlayerArrowSchema()
	
	// Test valid schema
	if err := schema.ValidateSchema(); err != nil {
		t.Errorf("Expected valid schema, got error: %v", err)
	}
	
	// Test schema with nil Arrow schema
	invalidSchema := &PlayerArrowSchema{schema: nil, version: CurrentSchemaVersion}
	if err := invalidSchema.ValidateSchema(); err == nil {
		t.Error("Expected error for nil schema, got nil")
	}
}

func TestGetFieldByName(t *testing.T) {
	schema := NewPlayerArrowSchema()
	
	// Test existing field
	field, found := schema.GetFieldByName("uid")
	if !found {
		t.Error("Expected to find 'uid' field")
	}
	if field.Name != "uid" {
		t.Errorf("Expected field name 'uid', got '%s'", field.Name)
	}
	
	// Test non-existing field
	_, found = schema.GetFieldByName("non_existent_field")
	if found {
		t.Error("Expected not to find 'non_existent_field'")
	}
}

func TestIsCompatibleWith(t *testing.T) {
	schema := NewPlayerArrowSchema()
	
	// Test compatibility with same version
	if !schema.IsCompatibleWith(CurrentSchemaVersion) {
		t.Error("Expected schema to be compatible with current version")
	}
	
	// Test compatibility with different version
	if schema.IsCompatibleWith("v2.0.0") {
		t.Error("Expected schema to be incompatible with different version")
	}
}

func TestCreateDatasetMetadataSchema(t *testing.T) {
	schema := CreateDatasetMetadataSchema()
	
	if schema == nil {
		t.Fatal("Expected metadata schema to be created, got nil")
	}
	
	// Test required fields exist
	requiredFields := []string{
		"schema_version", "created_at", "updated_at", "player_count",
		"source_format", "currency_symbol", "dataset_id",
	}
	
	for _, fieldName := range requiredFields {
		found := false
		for _, field := range schema.Fields() {
			if field.Name == fieldName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected metadata field '%s' to exist", fieldName)
		}
	}
	
	// Test metadata
	metadata := schema.Metadata()
	if metadata.Len() == 0 {
		t.Fatal("Expected metadata schema to have metadata")
	}
	
	// Check for schema_version
	versionFound := false
	for i := 0; i < metadata.Len(); i++ {
		if metadata.Keys()[i] == "schema_version" {
			versionFound = true
			break
		}
	}
	if !versionFound {
		t.Error("Expected metadata schema to have schema_version metadata")
	}
	
	// Check for schema_type
	typeFound := false
	for i := 0; i < metadata.Len(); i++ {
		if metadata.Keys()[i] == "schema_type" {
			if metadata.Values()[i] != "dataset_metadata" {
				t.Errorf("Expected schema_type 'dataset_metadata', got '%s'", metadata.Values()[i])
			}
			typeFound = true
			break
		}
	}
	if !typeFound {
		t.Error("Expected metadata schema to have schema_type metadata")
	}
}

func TestSchemaEvolution(t *testing.T) {
	allocator := memory.NewGoAllocator()
	evolution := NewSchemaEvolution(allocator)
	
	if evolution == nil {
		t.Fatal("Expected schema evolution to be created, got nil")
	}
	
	currentSchema := evolution.GetCurrentSchema()
	if currentSchema == nil {
		t.Fatal("Expected current schema to exist, got nil")
	}
	
	if currentSchema.GetVersion() != CurrentSchemaVersion {
		t.Errorf("Expected current schema version %s, got %s", CurrentSchemaVersion, currentSchema.GetVersion())
	}
	
	// Test migration capability
	if !evolution.CanMigrate(CurrentSchemaVersion, CurrentSchemaVersion) {
		t.Error("Expected to be able to migrate from current version to current version")
	}
	
	if evolution.CanMigrate(CurrentSchemaVersion, "v2.0.0") {
		t.Error("Expected not to be able to migrate to unsupported version")
	}
	
	// Test migration plan
	plan, err := evolution.GetMigrationPlan(CurrentSchemaVersion, CurrentSchemaVersion)
	if err != nil {
		t.Errorf("Expected no error for same version migration plan, got: %v", err)
	}
	if len(plan) != 1 || plan[0] != "no_migration_needed" {
		t.Errorf("Expected migration plan ['no_migration_needed'], got %v", plan)
	}
	
	// Test unsupported migration plan
	_, err = evolution.GetMigrationPlan(CurrentSchemaVersion, "v2.0.0")
	if err == nil {
		t.Error("Expected error for unsupported migration plan, got nil")
	}
}

func BenchmarkSchemaCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewPlayerArrowSchema()
	}
}

func BenchmarkSchemaValidation(b *testing.B) {
	schema := NewPlayerArrowSchema()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_ = schema.ValidateSchema()
	}
}

func BenchmarkFieldLookup(b *testing.B) {
	schema := NewPlayerArrowSchema()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, _ = schema.GetFieldByName("uid")
	}
}