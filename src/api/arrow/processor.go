package arrow

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/apache/arrow/go/v17/arrow"
	"github.com/apache/arrow/go/v17/arrow/array"
	"github.com/apache/arrow/go/v17/arrow/memory"
)

// ArrowProcessor defines the interface for Arrow data processing operations
type ArrowProcessor interface {
	ConvertFromJSON(ctx context.Context, players []Player) (arrow.Table, error)
	ConvertToJSON(ctx context.Context, table arrow.Table) ([]Player, error)
	Filter(ctx context.Context, table arrow.Table, filters []FilterExpression) (arrow.Table, error)
	Aggregate(ctx context.Context, table arrow.Table, aggregations []AggregationExpression) (arrow.Table, error)
	Sort(ctx context.Context, table arrow.Table, sortColumns []SortColumn) (arrow.Table, error)
	Project(ctx context.Context, table arrow.Table, columns []string) (arrow.Table, error)
}

// RoleOverallScore stores the calculated overall score for a player in a specific role.
type RoleOverallScore struct {
	RoleName string `json:"roleName"`
	Score    int    `json:"score"`
}

// Player holds all the information and calculated statistics for a football player.
type Player struct {
	UID                     int64                         `json:"uid"`
	Name                    string                        `json:"name"`
	Position                string                        `json:"position"`
	Age                     string                        `json:"age"`
	Club                    string                        `json:"club"`
	Division                string                        `json:"division"`
	TransferValue           string                        `json:"transfer_value"`
	Wage                    string                        `json:"wage"`
	Personality             string                        `json:"personality,omitempty"`
	MediaHandling           string                        `json:"media_handling,omitempty"`
	Nationality             string                        `json:"nationality"`
	NationalityISO          string                        `json:"nationality_iso"`
	NationalityFIFACode     string                        `json:"nationality_fifa_code"`
	AttributeMasked         bool                          `json:"attributeMasked,omitempty"`
	Attributes              map[string]string             `json:"attributes"`
	NumericAttributes       map[string]int                `json:"numericAttributes"`
	PerformanceStatsNumeric map[string]float64            `json:"performanceStatsNumeric"`
	PerformancePercentiles  map[string]map[string]float64 `json:"performancePercentiles"`
	ParsedPositions         []string                      `json:"parsedPositions"`
	ShortPositions          []string                      `json:"shortPositions"`
	PositionGroups          []string                      `json:"positionGroups"`
	PAC                     int                           `json:"PAC"`
	SHO                     int                           `json:"SHO"`
	PAS                     int                           `json:"PAS"`
	DRI                     int                           `json:"DRI"`
	DEF                     int                           `json:"DEF"`
	PHY                     int                           `json:"PHY"`
	GK                      int                           `json:"GK,omitempty"`
	DIV                     int                           `json:"DIV,omitempty"`
	HAN                     int                           `json:"HAN,omitempty"`
	REF                     int                           `json:"REF,omitempty"`
	KIC                     int                           `json:"KIC,omitempty"`
	SPD                     int                           `json:"SPD,omitempty"`
	POS                     int                           `json:"POS,omitempty"`
	Overall                 int                           `json:"Overall"`
	BestRoleOverall         string                        `json:"bestRoleOverall"`
	RoleSpecificOveralls    []RoleOverallScore            `json:"roleSpecificOveralls"`
	TransferValueAmount     int64                         `json:"transferValueAmount"`
	WageAmount              int64                         `json:"wageAmount"`
}

// FilterExpression represents a filter condition
type FilterExpression struct {
	Column    string          `json:"column"`
	Operator  FilterOperator  `json:"operator"`
	Value     interface{}     `json:"value"`
	LogicalOp LogicalOperator `json:"logical_op,omitempty"`
}

// FilterOperator represents filter operators
type FilterOperator int

const (
	FilterEqual FilterOperator = iota
	FilterNotEqual
	FilterGreaterThan
	FilterLessThan
	FilterGreaterThanOrEqual
	FilterLessThanOrEqual
	FilterIn
	FilterNotIn
	FilterLike
	FilterNotLike
)

// LogicalOperator represents logical operators for combining filters
type LogicalOperator int

const (
	LogicalAnd LogicalOperator = iota
	LogicalOr
)

// AggregationExpression represents an aggregation operation
type AggregationExpression struct {
	Column   string            `json:"column"`
	Function AggregateFunction `json:"function"`
	Alias    string            `json:"alias"`
}

// AggregateFunction represents aggregation functions
type AggregateFunction int

const (
	AggSum AggregateFunction = iota
	AggCount
	AggAvg
	AggMin
	AggMax
	AggStdDev
	AggVariance
)

// SortColumn represents a column to sort by
type SortColumn struct {
	Column    string    `json:"column"`
	Direction SortOrder `json:"direction"`
}

// SortOrder represents sort direction
type SortOrder int

const (
	SortAsc SortOrder = iota
	SortDesc
)

// ArrowProcessorImpl implements ArrowProcessor interface
type ArrowProcessorImpl struct {
	memoryPool memory.Allocator
}

// NewArrowProcessor creates a new ArrowProcessor instance
func NewArrowProcessor() *ArrowProcessorImpl {
	return &ArrowProcessorImpl{
		memoryPool: memory.NewGoAllocator(),
	}
}

// ConvertFromJSON converts a slice of Player structs to an Arrow table
func (ap *ArrowProcessorImpl) ConvertFromJSON(ctx context.Context, players []Player) (arrow.Table, error) {
	if len(players) == 0 {
		// Create empty columns for each field in the schema
		columns := make([]arrow.Column, len(PlayerSchema.Fields()))
		for i, field := range PlayerSchema.Fields() {
			builder := array.NewBuilder(ap.memoryPool, field.Type)
			arr := builder.NewArray()
			chunked := arrow.NewChunked(field.Type, []arrow.Array{arr})
			columns[i] = *arrow.NewColumn(field, chunked)
			builder.Release()
			arr.Release()
		}
		return array.NewTable(PlayerSchema, columns, 0), nil
	}

	// Create builders for each field in the schema
	builders := make([]array.Builder, len(PlayerSchema.Fields()))
	for i, field := range PlayerSchema.Fields() {
		builders[i] = array.NewBuilder(ap.memoryPool, field.Type)
	}

	// Process each player
	for _, player := range players {
		if err := ap.appendPlayerToBuilders(builders, player); err != nil {
			// Clean up builders on error
			for _, builder := range builders {
				builder.Release()
			}
			return nil, fmt.Errorf("failed to convert player %d: %w", player.UID, err)
		}
	}

	// Build arrays from builders
	columns := make([]arrow.Column, len(builders))
	for i, builder := range builders {
		arr := builder.NewArray()
		chunked := arrow.NewChunked(PlayerSchema.Field(i).Type, []arrow.Array{arr})
		columns[i] = *arrow.NewColumn(PlayerSchema.Field(i), chunked)
		builder.Release()
		arr.Release()
	}

	table := array.NewTable(PlayerSchema, columns, int64(len(players)))
	return table, nil
}

// appendPlayerToBuilders appends a single player's data to the array builders
func (ap *ArrowProcessorImpl) appendPlayerToBuilders(builders []array.Builder, player Player) error {
	fieldIndex := 0

	// Helper function to get next builder
	nextBuilder := func() array.Builder {
		builder := builders[fieldIndex]
		fieldIndex++
		return builder
	}

	// uid (int64)
	nextBuilder().(*array.Int64Builder).Append(player.UID)

	// name (string)
	nextBuilder().(*array.StringBuilder).Append(player.Name)

	// position (string)
	nextBuilder().(*array.StringBuilder).Append(player.Position)

	// age (string)
	nextBuilder().(*array.StringBuilder).Append(player.Age)

	// club (string)
	nextBuilder().(*array.StringBuilder).Append(player.Club)

	// division (string)
	nextBuilder().(*array.StringBuilder).Append(player.Division)

	// nationality (string)
	nextBuilder().(*array.StringBuilder).Append(player.Nationality)

	// nationality_iso (string)
	nextBuilder().(*array.StringBuilder).Append(player.NationalityISO)

	// nationality_fifa_code (string)
	nextBuilder().(*array.StringBuilder).Append(player.NationalityFIFACode)

	// transfer_value (string)
	nextBuilder().(*array.StringBuilder).Append(player.TransferValue)

	// wage (string)
	nextBuilder().(*array.StringBuilder).Append(player.Wage)

	// transfer_value_amount (int64)
	nextBuilder().(*array.Int64Builder).Append(player.TransferValueAmount)

	// wage_amount (int64)
	nextBuilder().(*array.Int64Builder).Append(player.WageAmount)

	// personality (nullable string)
	personalityBuilder := nextBuilder().(*array.StringBuilder)
	if player.Personality != "" {
		personalityBuilder.Append(player.Personality)
	} else {
		personalityBuilder.AppendNull()
	}

	// media_handling (nullable string)
	mediaBuilder := nextBuilder().(*array.StringBuilder)
	if player.MediaHandling != "" {
		mediaBuilder.Append(player.MediaHandling)
	} else {
		mediaBuilder.AppendNull()
	}

	// attribute_masked (nullable bool)
	attrMaskedBuilder := nextBuilder().(*array.BooleanBuilder)
	if player.AttributeMasked {
		attrMaskedBuilder.Append(player.AttributeMasked)
	} else {
		attrMaskedBuilder.AppendNull()
	}

	// Core attributes (int32)
	nextBuilder().(*array.Int32Builder).Append(int32(player.PAC))
	nextBuilder().(*array.Int32Builder).Append(int32(player.SHO))
	nextBuilder().(*array.Int32Builder).Append(int32(player.PAS))
	nextBuilder().(*array.Int32Builder).Append(int32(player.DRI))
	nextBuilder().(*array.Int32Builder).Append(int32(player.DEF))
	nextBuilder().(*array.Int32Builder).Append(int32(player.PHY))
	nextBuilder().(*array.Int32Builder).Append(int32(player.Overall))

	// Goalkeeper attributes (nullable int32)
	gkBuilder := nextBuilder().(*array.Int32Builder)
	if player.GK > 0 {
		gkBuilder.Append(int32(player.GK))
	} else {
		gkBuilder.AppendNull()
	}

	divBuilder := nextBuilder().(*array.Int32Builder)
	if player.DIV > 0 {
		divBuilder.Append(int32(player.DIV))
	} else {
		divBuilder.AppendNull()
	}

	hanBuilder := nextBuilder().(*array.Int32Builder)
	if player.HAN > 0 {
		hanBuilder.Append(int32(player.HAN))
	} else {
		hanBuilder.AppendNull()
	}

	refBuilder := nextBuilder().(*array.Int32Builder)
	if player.REF > 0 {
		refBuilder.Append(int32(player.REF))
	} else {
		refBuilder.AppendNull()
	}

	kicBuilder := nextBuilder().(*array.Int32Builder)
	if player.KIC > 0 {
		kicBuilder.Append(int32(player.KIC))
	} else {
		kicBuilder.AppendNull()
	}

	spdBuilder := nextBuilder().(*array.Int32Builder)
	if player.SPD > 0 {
		spdBuilder.Append(int32(player.SPD))
	} else {
		spdBuilder.AppendNull()
	}

	posBuilder := nextBuilder().(*array.Int32Builder)
	if player.POS > 0 {
		posBuilder.Append(int32(player.POS))
	} else {
		posBuilder.AppendNull()
	}

	// best_role_overall (string)
	nextBuilder().(*array.StringBuilder).Append(player.BestRoleOverall)

	// Complex nested data - serialize to JSON strings
	// attributes (JSON string)
	attributesJSON, err := json.Marshal(player.Attributes)
	if err != nil {
		return fmt.Errorf("failed to marshal attributes: %w", err)
	}
	nextBuilder().(*array.StringBuilder).Append(string(attributesJSON))

	// numeric_attributes (JSON string)
	numericAttrsJSON, err := json.Marshal(player.NumericAttributes)
	if err != nil {
		return fmt.Errorf("failed to marshal numeric attributes: %w", err)
	}
	nextBuilder().(*array.StringBuilder).Append(string(numericAttrsJSON))

	// performance_stats_numeric (JSON string)
	perfStatsJSON, err := json.Marshal(player.PerformanceStatsNumeric)
	if err != nil {
		return fmt.Errorf("failed to marshal performance stats: %w", err)
	}
	nextBuilder().(*array.StringBuilder).Append(string(perfStatsJSON))

	// performance_percentiles (JSON string)
	perfPercentilesJSON, err := json.Marshal(player.PerformancePercentiles)
	if err != nil {
		return fmt.Errorf("failed to marshal performance percentiles: %w", err)
	}
	nextBuilder().(*array.StringBuilder).Append(string(perfPercentilesJSON))

	// parsed_positions (JSON string)
	parsedPosJSON, err := json.Marshal(player.ParsedPositions)
	if err != nil {
		return fmt.Errorf("failed to marshal parsed positions: %w", err)
	}
	nextBuilder().(*array.StringBuilder).Append(string(parsedPosJSON))

	// short_positions (JSON string)
	shortPosJSON, err := json.Marshal(player.ShortPositions)
	if err != nil {
		return fmt.Errorf("failed to marshal short positions: %w", err)
	}
	nextBuilder().(*array.StringBuilder).Append(string(shortPosJSON))

	// position_groups (JSON string)
	posGroupsJSON, err := json.Marshal(player.PositionGroups)
	if err != nil {
		return fmt.Errorf("failed to marshal position groups: %w", err)
	}
	nextBuilder().(*array.StringBuilder).Append(string(posGroupsJSON))

	// role_specific_overalls (JSON string)
	roleOverallsJSON, err := json.Marshal(player.RoleSpecificOveralls)
	if err != nil {
		return fmt.Errorf("failed to marshal role specific overalls: %w", err)
	}
	nextBuilder().(*array.StringBuilder).Append(string(roleOverallsJSON))

	return nil
}

// ConvertToJSON converts an Arrow table to a slice of Player structs
func (ap *ArrowProcessorImpl) ConvertToJSON(ctx context.Context, table arrow.Table) ([]Player, error) {
	if table == nil {
		return nil, fmt.Errorf("table cannot be nil")
	}

	numRows := int(table.NumRows())
	if numRows == 0 {
		return []Player{}, nil
	}

	players := make([]Player, numRows)

	// Process each row
	for rowIdx := 0; rowIdx < numRows; rowIdx++ {
		player, err := ap.extractPlayerFromRow(table, rowIdx)
		if err != nil {
			return nil, fmt.Errorf("failed to extract player at row %d: %w", rowIdx, err)
		}
		players[rowIdx] = player
	}

	return players, nil
}

// extractPlayerFromRow extracts a single Player struct from a table row
func (ap *ArrowProcessorImpl) extractPlayerFromRow(table arrow.Table, rowIdx int) (Player, error) {
	var player Player
	fieldIndex := 0

	// Helper function to get column value safely
	getColumnValue := func(colName string) (interface{}, error) {
		if fieldIndex >= int(table.NumCols()) {
			return nil, fmt.Errorf("field index %d exceeds number of columns %d", fieldIndex, table.NumCols())
		}
		
		column := table.Column(fieldIndex)
		fieldIndex++
		
		if column.Len() <= rowIdx {
			return nil, fmt.Errorf("row index %d exceeds column length %d", rowIdx, column.Len())
		}

		chunk := column.Data().Chunk(0) // Assuming single chunk for simplicity
		if chunk.IsNull(rowIdx) {
			return nil, nil
		}

		switch arr := chunk.(type) {
		case *array.Int64:
			return arr.Value(rowIdx), nil
		case *array.Int32:
			return arr.Value(rowIdx), nil
		case *array.String:
			return arr.Value(rowIdx), nil
		case *array.Boolean:
			return arr.Value(rowIdx), nil
		default:
			return nil, fmt.Errorf("unsupported array type for column %s: %T", colName, arr)
		}
	}

	// Extract fields in schema order
	// uid (int64)
	if val, err := getColumnValue("uid"); err != nil {
		return player, err
	} else if val != nil {
		player.UID = val.(int64)
	}

	// name (string)
	if val, err := getColumnValue("name"); err != nil {
		return player, err
	} else if val != nil {
		player.Name = val.(string)
	}

	// position (string)
	if val, err := getColumnValue("position"); err != nil {
		return player, err
	} else if val != nil {
		player.Position = val.(string)
	}

	// age (string)
	if val, err := getColumnValue("age"); err != nil {
		return player, err
	} else if val != nil {
		player.Age = val.(string)
	}

	// club (string)
	if val, err := getColumnValue("club"); err != nil {
		return player, err
	} else if val != nil {
		player.Club = val.(string)
	}

	// division (string)
	if val, err := getColumnValue("division"); err != nil {
		return player, err
	} else if val != nil {
		player.Division = val.(string)
	}

	// nationality (string)
	if val, err := getColumnValue("nationality"); err != nil {
		return player, err
	} else if val != nil {
		player.Nationality = val.(string)
	}

	// nationality_iso (string)
	if val, err := getColumnValue("nationality_iso"); err != nil {
		return player, err
	} else if val != nil {
		player.NationalityISO = val.(string)
	}

	// nationality_fifa_code (string)
	if val, err := getColumnValue("nationality_fifa_code"); err != nil {
		return player, err
	} else if val != nil {
		player.NationalityFIFACode = val.(string)
	}

	// transfer_value (string)
	if val, err := getColumnValue("transfer_value"); err != nil {
		return player, err
	} else if val != nil {
		player.TransferValue = val.(string)
	}

	// wage (string)
	if val, err := getColumnValue("wage"); err != nil {
		return player, err
	} else if val != nil {
		player.Wage = val.(string)
	}

	// transfer_value_amount (int64)
	if val, err := getColumnValue("transfer_value_amount"); err != nil {
		return player, err
	} else if val != nil {
		player.TransferValueAmount = val.(int64)
	}

	// wage_amount (int64)
	if val, err := getColumnValue("wage_amount"); err != nil {
		return player, err
	} else if val != nil {
		player.WageAmount = val.(int64)
	}

	// personality (nullable string)
	if val, err := getColumnValue("personality"); err != nil {
		return player, err
	} else if val != nil {
		player.Personality = val.(string)
	}

	// media_handling (nullable string)
	if val, err := getColumnValue("media_handling"); err != nil {
		return player, err
	} else if val != nil {
		player.MediaHandling = val.(string)
	}

	// attribute_masked (nullable bool)
	if val, err := getColumnValue("attribute_masked"); err != nil {
		return player, err
	} else if val != nil {
		player.AttributeMasked = val.(bool)
	}

	// Core attributes (int32)
	if val, err := getColumnValue("pac"); err != nil {
		return player, err
	} else if val != nil {
		player.PAC = int(val.(int32))
	}

	if val, err := getColumnValue("sho"); err != nil {
		return player, err
	} else if val != nil {
		player.SHO = int(val.(int32))
	}

	if val, err := getColumnValue("pas"); err != nil {
		return player, err
	} else if val != nil {
		player.PAS = int(val.(int32))
	}

	if val, err := getColumnValue("dri"); err != nil {
		return player, err
	} else if val != nil {
		player.DRI = int(val.(int32))
	}

	if val, err := getColumnValue("def"); err != nil {
		return player, err
	} else if val != nil {
		player.DEF = int(val.(int32))
	}

	if val, err := getColumnValue("phy"); err != nil {
		return player, err
	} else if val != nil {
		player.PHY = int(val.(int32))
	}

	if val, err := getColumnValue("overall"); err != nil {
		return player, err
	} else if val != nil {
		player.Overall = int(val.(int32))
	}

	// Goalkeeper attributes (nullable int32)
	if val, err := getColumnValue("gk"); err != nil {
		return player, err
	} else if val != nil {
		player.GK = int(val.(int32))
	}

	if val, err := getColumnValue("div"); err != nil {
		return player, err
	} else if val != nil {
		player.DIV = int(val.(int32))
	}

	if val, err := getColumnValue("han"); err != nil {
		return player, err
	} else if val != nil {
		player.HAN = int(val.(int32))
	}

	if val, err := getColumnValue("ref"); err != nil {
		return player, err
	} else if val != nil {
		player.REF = int(val.(int32))
	}

	if val, err := getColumnValue("kic"); err != nil {
		return player, err
	} else if val != nil {
		player.KIC = int(val.(int32))
	}

	if val, err := getColumnValue("spd"); err != nil {
		return player, err
	} else if val != nil {
		player.SPD = int(val.(int32))
	}

	if val, err := getColumnValue("pos"); err != nil {
		return player, err
	} else if val != nil {
		player.POS = int(val.(int32))
	}

	// best_role_overall (string)
	if val, err := getColumnValue("best_role_overall"); err != nil {
		return player, err
	} else if val != nil {
		player.BestRoleOverall = val.(string)
	}

	// Complex nested data - deserialize from JSON strings
	// attributes
	if val, err := getColumnValue("attributes"); err != nil {
		return player, err
	} else if val != nil {
		if err := json.Unmarshal([]byte(val.(string)), &player.Attributes); err != nil {
			return player, fmt.Errorf("failed to unmarshal attributes: %w", err)
		}
	}

	// numeric_attributes
	if val, err := getColumnValue("numeric_attributes"); err != nil {
		return player, err
	} else if val != nil {
		if err := json.Unmarshal([]byte(val.(string)), &player.NumericAttributes); err != nil {
			return player, fmt.Errorf("failed to unmarshal numeric attributes: %w", err)
		}
	}

	// performance_stats_numeric
	if val, err := getColumnValue("performance_stats_numeric"); err != nil {
		return player, err
	} else if val != nil {
		if err := json.Unmarshal([]byte(val.(string)), &player.PerformanceStatsNumeric); err != nil {
			return player, fmt.Errorf("failed to unmarshal performance stats: %w", err)
		}
	}

	// performance_percentiles
	if val, err := getColumnValue("performance_percentiles"); err != nil {
		return player, err
	} else if val != nil {
		if err := json.Unmarshal([]byte(val.(string)), &player.PerformancePercentiles); err != nil {
			return player, fmt.Errorf("failed to unmarshal performance percentiles: %w", err)
		}
	}

	// parsed_positions
	if val, err := getColumnValue("parsed_positions"); err != nil {
		return player, err
	} else if val != nil {
		if err := json.Unmarshal([]byte(val.(string)), &player.ParsedPositions); err != nil {
			return player, fmt.Errorf("failed to unmarshal parsed positions: %w", err)
		}
	}

	// short_positions
	if val, err := getColumnValue("short_positions"); err != nil {
		return player, err
	} else if val != nil {
		if err := json.Unmarshal([]byte(val.(string)), &player.ShortPositions); err != nil {
			return player, fmt.Errorf("failed to unmarshal short positions: %w", err)
		}
	}

	// position_groups
	if val, err := getColumnValue("position_groups"); err != nil {
		return player, err
	} else if val != nil {
		if err := json.Unmarshal([]byte(val.(string)), &player.PositionGroups); err != nil {
			return player, fmt.Errorf("failed to unmarshal position groups: %w", err)
		}
	}

	// role_specific_overalls
	if val, err := getColumnValue("role_specific_overalls"); err != nil {
		return player, err
	} else if val != nil {
		if err := json.Unmarshal([]byte(val.(string)), &player.RoleSpecificOveralls); err != nil {
			return player, fmt.Errorf("failed to unmarshal role specific overalls: %w", err)
		}
	}

	return player, nil
}

// Filter applies filter expressions to an Arrow table
func (ap *ArrowProcessorImpl) Filter(ctx context.Context, table arrow.Table, filters []FilterExpression) (arrow.Table, error) {
	// Implementation will be added in later tasks
	return nil, nil
}

// Aggregate applies aggregation expressions to an Arrow table
func (ap *ArrowProcessorImpl) Aggregate(ctx context.Context, table arrow.Table, aggregations []AggregationExpression) (arrow.Table, error) {
	// Implementation will be added in later tasks
	return nil, nil
}

// Sort sorts an Arrow table by specified columns
func (ap *ArrowProcessorImpl) Sort(ctx context.Context, table arrow.Table, sortColumns []SortColumn) (arrow.Table, error) {
	// Implementation will be added in later tasks
	return nil, nil
}

// Project selects specific columns from an Arrow table
func (ap *ArrowProcessorImpl) Project(ctx context.Context, table arrow.Table, columns []string) (arrow.Table, error) {
	// Implementation will be added in later tasks
	return nil, nil
}