package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// datasetQueryServiceURL is the base URL of the Dataset Query Service
// (src/dataset-query/), e.g. "http://localhost:8092". Empty (the default in
// every existing dev/test/CI setup) disables materialize-on-upload entirely
// -- there is deliberately no separate boolean "enabled" flag, since gating
// on URL-non-empty alone avoids an enabled-true-but-url-unset (or vice
// versa) contradictory state.
var datasetQueryServiceURL = strings.TrimRight(os.Getenv("DATASET_QUERY_SERVICE_URL"), "/")

// materializePlayerRow mirrors src/dataset-query/materialize_writer.go's
// PlayerRow field-for-field. Deliberately not shared/imported across the two
// modules -- they are separate, loosely-coupled services communicating over
// HTTP/JSON, same as every other phase of this migration.
type materializePlayerRow struct {
	PlayerID                int64              `json:"playerId"`
	Name                    string             `json:"name"`
	Position                string             `json:"position"`
	Age                     string             `json:"age"`
	Club                    string             `json:"club"`
	Division                string             `json:"division"`
	BasedIn                 string             `json:"basedIn"`
	TransferValue           string             `json:"transferValue"`
	TransferValueAmount     int64              `json:"transferValueAmount"`
	Wage                    string             `json:"wage"`
	WageAmount              int64              `json:"wageAmount"`
	Personality             string             `json:"personality"`
	MediaHandling           string             `json:"mediaHandling"`
	Nationality             string             `json:"nationality"`
	NationalityISO          string             `json:"nationalityIso"`
	NationalityFIFACode     string             `json:"nationalityFifaCode"`
	AttributeMasked         bool               `json:"attributeMasked"`
	NumericAttributes       map[string]int     `json:"numericAttributes"`
	PerformanceStatsNumeric map[string]float64 `json:"performanceStatsNumeric"`
	ParsedPositions         []string           `json:"parsedPositions"`
	ShortPositions          []string           `json:"shortPositions"`
	PositionGroups          []string           `json:"positionGroups"`
}

// materializeRequest is the POST /internal/materialize/{datasetId} body.
type materializeRequest struct {
	Players []materializePlayerRow `json:"players"`
}

// playerToMaterializeRow maps a Player (types.go) onto the shape the
// Dataset Query Service's materialize endpoint expects.
func playerToMaterializeRow(p Player) materializePlayerRow {
	return materializePlayerRow{
		PlayerID:                p.UID,
		Name:                    p.Name,
		Position:                p.Position,
		Age:                     p.Age,
		Club:                    p.Club,
		Division:                p.Division,
		BasedIn:                 p.BasedIn,
		TransferValue:           p.TransferValue,
		TransferValueAmount:     p.TransferValueAmount,
		Wage:                    p.Wage,
		WageAmount:              p.WageAmount,
		Personality:             p.Personality,
		MediaHandling:           p.MediaHandling,
		Nationality:             p.Nationality,
		NationalityISO:          p.NationalityISO,
		NationalityFIFACode:     p.NationalityFIFACode,
		AttributeMasked:         p.AttributeMasked,
		NumericAttributes:       p.NumericAttributes,
		PerformanceStatsNumeric: p.PerformanceStatsNumeric,
		ParsedPositions:         p.ParsedPositions,
		ShortPositions:          p.ShortPositions,
		PositionGroups:          p.PositionGroups,
	}
}

// materializeDataset POSTs players to the Dataset Query Service's
// POST /internal/materialize/{datasetId}, keeping its Parquet artifact for
// datasetID in sync with what src/api just stored. Synchronous and
// independently testable -- callers that want this to run in the
// background (e.g. uploadHandler, so it never blocks or fails the upload
// response) wrap the call in their own `go func() { ... }()` using
// context.Background(), not the inbound request's context, since that gets
// cancelled once the response is written.
//
// A no-op (returns nil immediately, no HTTP call attempted) whenever
// datasetQueryServiceURL is unset -- the default in every existing dev/
// test/CI setup today.
func materializeDataset(ctx context.Context, datasetID string, players []Player) error {
	if datasetQueryServiceURL == "" {
		return nil
	}

	rows := make([]materializePlayerRow, len(players))
	for i, p := range players {
		rows[i] = playerToMaterializeRow(p)
	}

	body, err := json.Marshal(materializeRequest{Players: rows})
	if err != nil {
		return fmt.Errorf("marshaling materialize request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/internal/materialize/%s", datasetQueryServiceURL, url.PathEscape(datasetID))
	resp, err := DefaultHTTPClient.Post(ctx, endpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("posting materialize request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("materialize request failed: status %d", resp.StatusCode)
	}
	return nil
}
