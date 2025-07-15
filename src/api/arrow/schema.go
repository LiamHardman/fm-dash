package arrow

import (
	"fmt"

	"github.com/apache/arrow/go/v17/arrow"
	"github.com/apache/arrow/go/v17/arrow/array"
	"github.com/apache/arrow/go/v17/arrow/memory"
)

// PlayerSchema defines the Arrow schema for Player data structure
var PlayerSchema = arrow.NewSchema([]arrow.Field{
	// Primary identifiers
	{Name: "uid", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
	{Name: "name", Type: arrow.BinaryTypes.String, Nullable: false},

	// Basic info (frequently filtered)
	{Name: "position", Type: arrow.BinaryTypes.String, Nullable: false},
	{Name: "age", Type: arrow.BinaryTypes.String, Nullable: false},
	{Name: "club", Type: arrow.BinaryTypes.String, Nullable: false},
	{Name: "division", Type: arrow.BinaryTypes.String, Nullable: false},
	{Name: "nationality", Type: arrow.BinaryTypes.String, Nullable: false},
	{Name: "nationality_iso", Type: arrow.BinaryTypes.String, Nullable: false},
	{Name: "nationality_fifa_code", Type: arrow.BinaryTypes.String, Nullable: false},

	// Transfer and wage info
	{Name: "transfer_value", Type: arrow.BinaryTypes.String, Nullable: false},
	{Name: "wage", Type: arrow.BinaryTypes.String, Nullable: false},
	{Name: "transfer_value_amount", Type: arrow.PrimitiveTypes.Int64, Nullable: false},
	{Name: "wage_amount", Type: arrow.PrimitiveTypes.Int64, Nullable: false},

	// Optional personality fields
	{Name: "personality", Type: arrow.BinaryTypes.String, Nullable: true},
	{Name: "media_handling", Type: arrow.BinaryTypes.String, Nullable: true},

	// Attribute masking flag
	{Name: "attribute_masked", Type: arrow.FixedWidthTypes.Boolean, Nullable: true},

	// Core attributes (frequently used in calculations)
	{Name: "pac", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
	{Name: "sho", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
	{Name: "pas", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
	{Name: "dri", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
	{Name: "def", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
	{Name: "phy", Type: arrow.PrimitiveTypes.Int32, Nullable: false},
	{Name: "overall", Type: arrow.PrimitiveTypes.Int32, Nullable: false},

	// Goalkeeper attributes (nullable for outfield players)
	{Name: "gk", Type: arrow.PrimitiveTypes.Int32, Nullable: true},
	{Name: "div", Type: arrow.PrimitiveTypes.Int32, Nullable: true},
	{Name: "han", Type: arrow.PrimitiveTypes.Int32, Nullable: true},
	{Name: "ref", Type: arrow.PrimitiveTypes.Int32, Nullable: true},
	{Name: "kic", Type: arrow.PrimitiveTypes.Int32, Nullable: true},
	{Name: "spd", Type: arrow.PrimitiveTypes.Int32, Nullable: true},
	{Name: "pos", Type: arrow.PrimitiveTypes.Int32, Nullable: true},

	// Best role information
	{Name: "best_role_overall", Type: arrow.BinaryTypes.String, Nullable: false},

	// Complex nested data - using JSON strings for now, can be optimized later
	{Name: "attributes", Type: arrow.BinaryTypes.String, Nullable: false},
	{Name: "numeric_attributes", Type: arrow.BinaryTypes.String, Nullable: false},
	{Name: "performance_stats_numeric", Type: arrow.BinaryTypes.String, Nullable: false},
	{Name: "performance_percentiles", Type: arrow.BinaryTypes.String, Nullable: false},
	{Name: "parsed_positions", Type: arrow.BinaryTypes.String, Nullable: false},
	{Name: "short_positions", Type: arrow.BinaryTypes.String, Nullable: false},
	{Name: "position_groups", Type: arrow.BinaryTypes.String, Nullable: false},
	{Name: "role_specific_overalls", Type: arrow.BinaryTypes.String, Nullable: false},
}, nil)

// SchemaValidator provides validation functions for Arrow schemas
type SchemaValidator struct {
	pool memory.Allocator
}

// NewSchemaValidator creates a new schema validator
func NewSchemaValidator() *SchemaValidator {
	return &SchemaValidator{
		pool: memory.NewGoAllocator(),
	}
}

// ValidateSchema validates an Arrow schema against expected structure
func (sv *SchemaValidator) ValidateSchema(schema *arrow.Schema) error {
	if schema == nil {
		return fmt.Errorf("schema cannot be nil")
	}

	// Check if schema has minimum required fields
	requiredFields := []string{
		"uid", "name", "position", "age", "club", "division",
		"nationality", "nationality_iso", "transfer_value_amount",
		"wage_amount", "pac", "sho", "pas", "dri", "def", "phy", "overall",
	}

	schemaFields := make(map[string]arrow.Field)
	for i, field := range schema.Fields() {
		schemaFields[field.Name] = schema.Field(i)
	}

	for _, required := range requiredFields {
		if _, exists := schemaFields[required]; !exists {
			return fmt.Errorf("required field '%s' is missing from schema", required)
		}
	}

	return nil
}

// ValidateSchemaCompatibility checks if two schemas are compatible
func (sv *SchemaValidator) ValidateSchemaCompatibility(oldSchema, newSchema *arrow.Schema) error {
	if oldSchema == nil || newSchema == nil {
		return fmt.Errorf("schemas cannot be nil")
	}

	// Check if all fields from old schema exist in new schema with compatible types
	oldFields := make(map[string]arrow.Field)
	for i, field := range oldSchema.Fields() {
		oldFields[field.Name] = oldSchema.Field(i)
	}

	newFields := make(map[string]arrow.Field)
	for i, field := range newSchema.Fields() {
		newFields[field.Name] = newSchema.Field(i)
	}

	for fieldName, oldField := range oldFields {
		newField, exists := newFields[fieldName]
		if !exists {
			return fmt.Errorf("field '%s' exists in old schema but not in new schema", fieldName)
		}

		// Check type compatibility
		if !sv.areTypesCompatible(oldField.Type, newField.Type) {
			return fmt.Errorf("field '%s' has incompatible type change: %s -> %s", 
				fieldName, oldField.Type.String(), newField.Type.String())
		}

		// Check nullability - can't make non-nullable field nullable without migration
		if !oldField.Nullable && newField.Nullable {
			// This is generally safe - making a field nullable
		} else if oldField.Nullable && !newField.Nullable {
			return fmt.Errorf("field '%s' cannot change from nullable to non-nullable without data migration", fieldName)
		}
	}

	return nil
}

// areTypesCompatible checks if two Arrow types are compatible
func (sv *SchemaValidator) areTypesCompatible(oldType, newType arrow.DataType) bool {
	// Exact type match
	if oldType.ID() == newType.ID() {
		return true
	}

	// Check for safe type promotions
	switch oldType.ID() {
	case arrow.INT32:
		// INT32 can be promoted to INT64
		return newType.ID() == arrow.INT64
	case arrow.FLOAT32:
		// FLOAT32 can be promoted to FLOAT64
		return newType.ID() == arrow.FLOAT64
	case arrow.STRING:
		// STRING is compatible with LARGE_STRING
		return newType.ID() == arrow.LARGE_STRING
	}

	return false
}

// CreateEmptyTable creates an empty Arrow table with the Player schema
func CreateEmptyTable() arrow.Table {
	pool := memory.NewGoAllocator()
	
	// Create empty arrays for each field
	builders := make([]array.Builder, len(PlayerSchema.Fields()))
	columns := make([]arrow.Column, len(PlayerSchema.Fields()))
	
	for i, field := range PlayerSchema.Fields() {
		builders[i] = array.NewBuilder(pool, field.Type)
		arr := builders[i].NewArray()
		chunked := arrow.NewChunked(field.Type, []arrow.Array{arr})
		columns[i] = *arrow.NewColumn(field, chunked)
		builders[i].Release()
		arr.Release()
	}
	
	table := array.NewTable(PlayerSchema, columns, 0)
	
	return table
}

// GetSchemaVersion returns the current schema version
func GetSchemaVersion() int {
	return 1
}

// GetSchemaFingerprint returns a unique fingerprint for the schema
func GetSchemaFingerprint(schema *arrow.Schema) string {
	if schema == nil {
		return ""
	}
	
	// Create a simple fingerprint based on field names and types
	fingerprint := ""
	for _, field := range schema.Fields() {
		fingerprint += fmt.Sprintf("%s:%s:%t;", field.Name, field.Type.String(), field.Nullable)
	}
	
	return fingerprint
}

// CompareSchemas compares two schemas and returns detailed differences
func CompareSchemas(oldSchema, newSchema *arrow.Schema) *SchemaComparison {
	comparison := &SchemaComparison{
		Compatible: true,
		Changes:    make([]SchemaChange, 0),
	}

	if oldSchema == nil || newSchema == nil {
		comparison.Compatible = false
		comparison.Changes = append(comparison.Changes, SchemaChange{
			Type:        SchemaChangeInvalid,
			Description: "One or both schemas are nil",
		})
		return comparison
	}

	oldFields := make(map[string]arrow.Field)
	for i, field := range oldSchema.Fields() {
		oldFields[field.Name] = oldSchema.Field(i)
	}

	newFields := make(map[string]arrow.Field)
	for i, field := range newSchema.Fields() {
		newFields[field.Name] = newSchema.Field(i)
	}

	// Check for removed fields
	for fieldName := range oldFields {
		if _, exists := newFields[fieldName]; !exists {
			comparison.Compatible = false
			comparison.Changes = append(comparison.Changes, SchemaChange{
				Type:        SchemaChangeRemoved,
				FieldName:   fieldName,
				Description: fmt.Sprintf("Field '%s' was removed", fieldName),
			})
		}
	}

	// Check for added fields and modified fields
	for fieldName, newField := range newFields {
		if oldField, exists := oldFields[fieldName]; exists {
			// Field exists in both schemas - check for modifications
			if oldField.Type.ID() != newField.Type.ID() {
				validator := NewSchemaValidator()
				if validator.areTypesCompatible(oldField.Type, newField.Type) {
					comparison.Changes = append(comparison.Changes, SchemaChange{
						Type:        SchemaChangeModified,
						FieldName:   fieldName,
						Description: fmt.Sprintf("Field '%s' type changed from %s to %s (compatible)", fieldName, oldField.Type.String(), newField.Type.String()),
					})
				} else {
					comparison.Compatible = false
					comparison.Changes = append(comparison.Changes, SchemaChange{
						Type:        SchemaChangeModified,
						FieldName:   fieldName,
						Description: fmt.Sprintf("Field '%s' type changed from %s to %s (incompatible)", fieldName, oldField.Type.String(), newField.Type.String()),
					})
				}
			}

			if oldField.Nullable != newField.Nullable {
				if oldField.Nullable && !newField.Nullable {
					comparison.Compatible = false
					comparison.Changes = append(comparison.Changes, SchemaChange{
						Type:        SchemaChangeModified,
						FieldName:   fieldName,
						Description: fmt.Sprintf("Field '%s' changed from nullable to non-nullable (requires migration)", fieldName),
					})
				} else {
					comparison.Changes = append(comparison.Changes, SchemaChange{
						Type:        SchemaChangeModified,
						FieldName:   fieldName,
						Description: fmt.Sprintf("Field '%s' changed from non-nullable to nullable", fieldName),
					})
				}
			}
		} else {
			// New field added
			comparison.Changes = append(comparison.Changes, SchemaChange{
				Type:        SchemaChangeAdded,
				FieldName:   fieldName,
				Description: fmt.Sprintf("Field '%s' was added", fieldName),
			})
		}
	}

	return comparison
}

// SchemaComparison represents the result of comparing two schemas
type SchemaComparison struct {
	Compatible bool
	Changes    []SchemaChange
}

// SchemaChange represents a single change between schemas
type SchemaChange struct {
	Type        SchemaChangeType
	FieldName   string
	Description string
}

// SchemaChangeType represents the type of schema change
type SchemaChangeType int

const (
	SchemaChangeAdded SchemaChangeType = iota
	SchemaChangeRemoved
	SchemaChangeModified
	SchemaChangeInvalid
)

func (sct SchemaChangeType) String() string {
	switch sct {
	case SchemaChangeAdded:
		return "ADDED"
	case SchemaChangeRemoved:
		return "REMOVED"
	case SchemaChangeModified:
		return "MODIFIED"
	case SchemaChangeInvalid:
		return "INVALID"
	default:
		return "UNKNOWN"
	}
}