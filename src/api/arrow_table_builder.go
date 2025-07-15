package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/apache/arrow/go/v18/arrow"
	"github.com/apache/arrow/go/v18/arrow/array"
	"github.com/apache/arrow/go/v18/arrow/memory"
)

// ArrowDatasetManager handles Arrow table operations and conversions
type ArrowDatasetManager struct {
	schema           *PlayerArrowSchema
	allocator        memory.Allocator
	builders         []array.Builder
	dictBuilders     map[string]*array.BinaryDictionaryBuilder
	stringInterner   *StringInterner
	recordBatchSize  int
}

// StringInterner manages dictionary encoding for repeated string values
type StringInterner struct {
	positionDict     map[string]uint8
	clubDict         map[string]uint16
	nationalityDict  map[string]uint8
	divisionDict     map[string]uint8
	nameDict         map[string]uint16
	nextPositionID   uint8
	nextClubID       uint16
	nextNationalityID uint8
	nextDivisionID   uint8
	nextNameID       uint16
}

// NewStringInterner creates a new string interner for dictionary encoding
func NewStringInterner() *StringInterner {
	return &StringInterner{
		positionDict:    make(map[string]uint8),
		clubDict:        make(map[string]uint16),
		nationalityDict: make(map[string]uint8),
		divisionDict:    make(map[string]uint8),
		nameDict:        make(map[string]uint16),
	}
}

// InternPosition returns the dictionary ID for a position string
func (si *StringInterner) InternPosition(position string) uint8 {
	if id, exists := si.positionDict[position]; exists {
		return id
	}
	id := si.nextPositionID
	si.positionDict[position] = id
	si.nextPositionID++
	return id
}

// InternClub returns the dictionary ID for a club string
func (si *StringInterner) InternClub(club string) uint16 {
	if id, exists := si.clubDict[club]; exists {
		return id
	}
	id := si.nextClubID
	si.clubDict[club] = id
	si.nextClubID++
	return id
}

// InternNationality returns the dictionary ID for a nationality string
func (si *StringInterner) InternNationality(nationality string) uint8 {
	if id, exists := si.nationalityDict[nationality]; exists {
		return id
	}
	id := si.nextNationalityID
	si.nationalityDict[nationality] = id
	si.nextNationalityID++
	return id
}

// InternDivision returns the dictionary ID for a division string
func (si *StringInterner) InternDivision(division string) uint8 {
	if id, exists := si.divisionDict[division]; exists {
		return id
	}
	id := si.nextDivisionID
	si.divisionDict[division] = id
	si.nextDivisionID++
	return id
}

// InternName returns the dictionary ID for a name string
func (si *StringInterner) InternName(name string) uint16 {
	if id, exists := si.nameDict[name]; exists {
		return id
	}
	id := si.nextNameID
	si.nameDict[name] = id
	si.nextNameID++
	return id
}

// GetPositionDictionary returns the position dictionary
func (si *StringInterner) GetPositionDictionary() map[string]uint8 {
	return si.positionDict
}

// GetClubDictionary returns the club dictionary
func (si *StringInterner) GetClubDictionary() map[string]uint16 {
	return si.clubDict
}

// GetNationalityDictionary returns the nationality dictionary
func (si *StringInterner) GetNationalityDictionary() map[string]uint8 {
	return si.nationalityDict
}

// GetDivisionDictionary returns the division dictionary
func (si *StringInterner) GetDivisionDictionary() map[string]uint8 {
	return si.divisionDict
}

// GetNameDictionary returns the name dictionary
func (si *StringInterner) GetNameDictionary() map[string]uint16 {
	return si.nameDict
}

// NewArrowDatasetManager creates a new Arrow dataset manager
func NewArrowDatasetManager(allocator memory.Allocator) *ArrowDatasetManager {
	return &ArrowDatasetManager{
		schema:          NewPlayerArrowSchema(),
		allocator:       allocator,
		dictBuilders:    make(map[string]*array.BinaryDictionaryBuilder),
		stringInterner:  NewStringInterner(),
		recordBatchSize: 10000, // Default batch size
	}
}

// SetRecordBatchSize sets the record batch size for Arrow table building
func (adm *ArrowDatasetManager) SetRecordBatchSize(size int) {
	adm.recordBatchSize = size
}

// initializeBuilders creates and initializes all column builders
func (adm *ArrowDatasetManager) initializeBuilders(numRows int) error {
	arrowSchema := adm.schema.GetSchema()
	adm.builders = make([]array.Builder, len(arrowSchema.Fields()))
	
	for i, field := range arrowSchema.Fields() {
		builder, err := adm.createBuilderForField(field)
		if err != nil {
			return fmt.Errorf("failed to create builder for field %s: %w", field.Name, err)
		}
		adm.builders[i] = builder
	}
	
	return nil
}

// createBuilderForField creates an appropriate builder for a given field
func (adm *ArrowDatasetManager) createBuilderForField(field arrow.Field) (array.Builder, error) {
	switch field.Type.ID() {
	case arrow.INT64:
		return array.NewInt64Builder(adm.allocator), nil
	case arrow.INT32:
		return array.NewInt32Builder(adm.allocator), nil
	case arrow.INT16:
		return array.NewInt16Builder(adm.allocator), nil
	case arrow.BOOL:
		return array.NewBooleanBuilder(adm.allocator), nil
	case arrow.STRING:
		return array.NewStringBuilder(adm.allocator), nil
	case arrow.DICTIONARY:
		// For now, use string builder instead of dictionary builder
		return array.NewStringBuilder(adm.allocator), nil
	case arrow.LIST:
		listType := field.Type.(*arrow.ListType)
		return array.NewListBuilderWithField(adm.allocator, arrow.Field{Name: "item", Type: listType.Elem(), Nullable: true}), nil
	case arrow.STRUCT:
		structType := field.Type.(*arrow.StructType)
		return array.NewStructBuilder(adm.allocator, structType), nil
	case arrow.FLOAT64:
		return array.NewFloat64Builder(adm.allocator), nil
	default:
		return nil, fmt.Errorf("unsupported field type: %s", field.Type)
	}
}

// BuildArrowTable converts player data to Arrow table
func (adm *ArrowDatasetManager) BuildArrowTable(players []Player) (arrow.Table, error) {
	if len(players) == 0 {
		return nil, fmt.Errorf("no players provided")
	}
	
	// Initialize builders
	if err := adm.initializeBuilders(len(players)); err != nil {
		return nil, fmt.Errorf("failed to initialize builders: %w", err)
	}
	
	// Process players in batches
	var records []arrow.Record
	batchStart := 0
	
	for batchStart < len(players) {
		batchEnd := batchStart + adm.recordBatchSize
		if batchEnd > len(players) {
			batchEnd = len(players)
		}
		
		batch := players[batchStart:batchEnd]
		record, err := adm.buildRecordBatch(batch)
		if err != nil {
			return nil, fmt.Errorf("failed to build record batch: %w", err)
		}
		
		records = append(records, record)
		batchStart = batchEnd
	}
	
	// Create table from records
	table := array.NewTableFromRecords(adm.schema.GetSchema(), records)
	
	// Clean up records
	for _, record := range records {
		record.Release()
	}
	
	return table, nil
}

// buildRecordBatch builds a single record batch from a slice of players
func (adm *ArrowDatasetManager) buildRecordBatch(players []Player) (arrow.Record, error) {
	// Reset builders for new batch
	for _, builder := range adm.builders {
		builder.Release()
	}
	
	if err := adm.initializeBuilders(len(players)); err != nil {
		return nil, fmt.Errorf("failed to reinitialize builders: %w", err)
	}
	
	// Populate builders with player data
	for _, player := range players {
		if err := adm.appendPlayerRecord(player); err != nil {
			return nil, fmt.Errorf("failed to append player record: %w", err)
		}
	}
	
	// Build arrays from builders
	arrays, err := adm.buildArrays()
	if err != nil {
		return nil, fmt.Errorf("failed to build arrays: %w", err)
	}
	
	// Create record from arrays
	record := array.NewRecord(adm.schema.GetSchema(), arrays, int64(len(players)))
	
	return record, nil
}

// appendPlayerRecord appends a single player record to the builders
func (adm *ArrowDatasetManager) appendPlayerRecord(player Player) error {
	arrowSchema := adm.schema.GetSchema()
	
	for i, field := range arrowSchema.Fields() {
		builder := adm.builders[i]
		
		switch field.Name {
		case "uid":
			builder.(*array.Int64Builder).Append(player.UID)
		case "overall":
			builder.(*array.Int32Builder).Append(int32(player.Overall))
		case "age_numeric":
			age, err := adm.parseAge(player.Age)
			if err != nil {
				return fmt.Errorf("failed to parse age: %w", err)
			}
			builder.(*array.Int32Builder).Append(int32(age))
		case "transfer_value_amount":
			builder.(*array.Int64Builder).Append(player.TransferValueAmount)
		case "wage_amount":
			builder.(*array.Int64Builder).Append(player.WageAmount)
		case "name":
			stringBuilder := builder.(*array.StringBuilder)
			stringBuilder.Append(player.Name)
		case "position":
			stringBuilder := builder.(*array.StringBuilder)
			stringBuilder.Append(player.Position)
		case "club":
			stringBuilder := builder.(*array.StringBuilder)
			stringBuilder.Append(player.Club)
		case "division":
			stringBuilder := builder.(*array.StringBuilder)
			stringBuilder.Append(player.Division)
		case "nationality":
			stringBuilder := builder.(*array.StringBuilder)
			stringBuilder.Append(player.Nationality)
		case "nationality_iso":
			stringBuilder := builder.(*array.StringBuilder)
			stringBuilder.Append(player.NationalityISO)
		case "nationality_fifa_code":
			stringBuilder := builder.(*array.StringBuilder)
			stringBuilder.Append(player.NationalityFIFACode)
		case "pac":
			builder.(*array.Int16Builder).Append(int16(player.PAC))
		case "sho":
			builder.(*array.Int16Builder).Append(int16(player.SHO))
		case "pas":
			builder.(*array.Int16Builder).Append(int16(player.PAS))
		case "dri":
			builder.(*array.Int16Builder).Append(int16(player.DRI))
		case "def":
			builder.(*array.Int16Builder).Append(int16(player.DEF))
		case "phy":
			builder.(*array.Int16Builder).Append(int16(player.PHY))
		case "gk":
			if player.GK > 0 {
				builder.(*array.Int16Builder).Append(int16(player.GK))
			} else {
				builder.(*array.Int16Builder).AppendNull()
			}
		case "div":
			if player.DIV > 0 {
				builder.(*array.Int16Builder).Append(int16(player.DIV))
			} else {
				builder.(*array.Int16Builder).AppendNull()
			}
		case "han":
			if player.HAN > 0 {
				builder.(*array.Int16Builder).Append(int16(player.HAN))
			} else {
				builder.(*array.Int16Builder).AppendNull()
			}
		case "ref":
			if player.REF > 0 {
				builder.(*array.Int16Builder).Append(int16(player.REF))
			} else {
				builder.(*array.Int16Builder).AppendNull()
			}
		case "kic":
			if player.KIC > 0 {
				builder.(*array.Int16Builder).Append(int16(player.KIC))
			} else {
				builder.(*array.Int16Builder).AppendNull()
			}
		case "spd":
			if player.SPD > 0 {
				builder.(*array.Int16Builder).Append(int16(player.SPD))
			} else {
				builder.(*array.Int16Builder).AppendNull()
			}
		case "pos":
			if player.POS > 0 {
				builder.(*array.Int16Builder).Append(int16(player.POS))
			} else {
				builder.(*array.Int16Builder).AppendNull()
			}
		case "personality":
			if player.Personality != "" {
				builder.(*array.StringBuilder).Append(player.Personality)
			} else {
				builder.(*array.StringBuilder).AppendNull()
			}
		case "media_handling":
			if player.MediaHandling != "" {
				builder.(*array.StringBuilder).Append(player.MediaHandling)
			} else {
				builder.(*array.StringBuilder).AppendNull()
			}
		case "best_role_overall":
			if player.BestRoleOverall != "" {
				builder.(*array.StringBuilder).Append(player.BestRoleOverall)
			} else {
				builder.(*array.StringBuilder).AppendNull()
			}
		case "attribute_masked":
			builder.(*array.BooleanBuilder).Append(player.AttributeMasked)
		case "parsed_positions":
			if err := adm.appendStringList(builder.(*array.ListBuilder), player.ParsedPositions); err != nil {
				return fmt.Errorf("failed to append parsed_positions: %w", err)
			}
		case "short_positions":
			if err := adm.appendStringList(builder.(*array.ListBuilder), player.ShortPositions); err != nil {
				return fmt.Errorf("failed to append short_positions: %w", err)
			}
		case "position_groups":
			if err := adm.appendStringList(builder.(*array.ListBuilder), player.PositionGroups); err != nil {
				return fmt.Errorf("failed to append position_groups: %w", err)
			}
		case "performance_stats":
			if err := adm.appendPerformanceStats(builder.(*array.StructBuilder), player.PerformanceStatsNumeric); err != nil {
				return fmt.Errorf("failed to append performance_stats: %w", err)
			}
		case "role_specific_overalls":
			if err := adm.appendRoleOveralls(builder.(*array.ListBuilder), player.RoleSpecificOveralls); err != nil {
				return fmt.Errorf("failed to append role_specific_overalls: %w", err)
			}
		default:
			return fmt.Errorf("unknown field: %s", field.Name)
		}
	}
	
	return nil
}

// appendStringList appends a string slice to a list builder
func (adm *ArrowDatasetManager) appendStringList(listBuilder *array.ListBuilder, values []string) error {
	listBuilder.Append(true)
	elemBuilder := listBuilder.ValueBuilder().(*array.StringBuilder)
	
	for _, value := range values {
		elemBuilder.Append(value)
	}
	
	return nil
}

// appendPerformanceStats appends performance stats to a struct builder
func (adm *ArrowDatasetManager) appendPerformanceStats(structBuilder *array.StructBuilder, stats map[string]float64) error {
	structBuilder.Append(true)
	
	// Map stats to struct fields
	statFields := []string{"goals", "assists", "pass_completion", "shots_per_game", "tackles_per_game"}
	
	for i, fieldName := range statFields {
		fieldBuilder := structBuilder.FieldBuilder(i)
		floatBuilder := fieldBuilder.(*array.Float64Builder)
		if value, exists := stats[fieldName]; exists {
			floatBuilder.Append(value)
		} else {
			floatBuilder.AppendNull()
		}
	}
	
	return nil
}

// appendRoleOveralls appends role-specific overalls to a list builder
func (adm *ArrowDatasetManager) appendRoleOveralls(listBuilder *array.ListBuilder, roles []RoleOverallScore) error {
	listBuilder.Append(true)
	elemBuilder := listBuilder.ValueBuilder().(*array.StructBuilder)
	
	for _, role := range roles {
		elemBuilder.Append(true)
		
		// role_name field
		nameBuilder := elemBuilder.FieldBuilder(0).(*array.StringBuilder)
		nameBuilder.Append(role.RoleName)
		
		// score field
		scoreBuilder := elemBuilder.FieldBuilder(1).(*array.Int32Builder)
		scoreBuilder.Append(int32(role.Score))
	}
	
	return nil
}

// buildArrays builds arrays from all builders
func (adm *ArrowDatasetManager) buildArrays() ([]arrow.Array, error) {
	arrays := make([]arrow.Array, len(adm.builders))
	
	for i, builder := range adm.builders {
		array := builder.NewArray()
		arrays[i] = array
	}
	
	return arrays, nil
}

// parseAge parses age string to integer
func (adm *ArrowDatasetManager) parseAge(ageStr string) (int, error) {
	// Handle various age formats
	ageStr = strings.TrimSpace(ageStr)
	
	// Remove any non-numeric characters except for the first number
	parts := strings.Fields(ageStr)
	if len(parts) > 0 {
		// Try to parse the first part as age
		if age, err := strconv.Atoi(parts[0]); err == nil {
			return age, nil
		}
	}
	
	// Fallback: try to extract first number from string
	var numStr strings.Builder
	for _, char := range ageStr {
		if char >= '0' && char <= '9' {
			numStr.WriteRune(char)
		} else if numStr.Len() > 0 {
			break // Stop at first non-digit after we've started collecting digits
		}
	}
	
	if numStr.Len() > 0 {
		return strconv.Atoi(numStr.String())
	}
	
	return 0, fmt.Errorf("unable to parse age from: %s", ageStr)
}

// ConvertPlayersToArrowTable is a convenience function to convert players to Arrow table
func ConvertPlayersToArrowTable(players []Player, allocator memory.Allocator) (arrow.Table, error) {
	manager := NewArrowDatasetManager(allocator)
	return manager.BuildArrowTable(players)
}

// Release releases all resources held by the dataset manager
func (adm *ArrowDatasetManager) Release() {
	for _, builder := range adm.builders {
		if builder != nil {
			builder.Release()
		}
	}
	adm.builders = nil
	adm.dictBuilders = nil
}