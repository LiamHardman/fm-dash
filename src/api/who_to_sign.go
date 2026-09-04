package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Who to Sign — LLM scouting assistant. Implements the spec charted in
// .scratch/who-to-sign/ (Wayfinder map "Who to Sign — LLM Scouting Assistant").

const (
	whoToSignModel         = "gpt-5.6-luna"
	whoToSignMaxToolRounds = 5
	whoToSignFindPlayersN  = 15

	// whoToSignFeature tags this feature's spans/metrics/errors (tracing map ticket 01).
	whoToSignFeature = "who_to_sign"
)

// whoToSignAttributeShortCodes maps the long attribute names shown in the frontend's
// dropdown (ticket 04) to the short codes actually used as Player.NumericAttributes
// keys (see player_processing.go's isNumericTechnicalAttribute switch).
var whoToSignAttributeShortCodes = map[string]string{
	"Acceleration": "Acc", "Pace": "Pac", "Strength": "Str", "Stamina": "Sta",
	"Natural Fitness": "Nat", "Balance": "Bal", "Jumping Reach": "Jum", "Agility": "Agi",
	"Aggression": "Agg", "Anticipation": "Ant", "Bravery": "Bra", "Composure": "Cmp",
	"Concentration": "Cnt", "Decisions": "Dec", "Determination": "Det", "Flair": "Fla",
	"Leadership": "Ldr", "Off The Ball": "OtB", "Positioning": "Pos", "Team Work": "Tea",
	"Vision": "Vis", "Work Rate": "Wor",
	"Corners": "Cor", "Crossing": "Cro", "Dribbling": "Dri", "Finishing": "Fin",
	"First Touch": "Fir", "Free Kick Taking": "Fre", "Heading": "Hea", "Long Shots": "Lon",
	"Long Throws": "L Th", "Marking": "Mar", "Passing": "Pas", "Penalty Taking": "Pen",
	"Tackling": "Tck", "Technique": "Tec",
	"Aerial Reach": "Aer", "Command Of Area": "Cmd", "Communication": "Com",
	"Eccentricity": "Ecc", "Handling": "Han", "Kicking": "Kic", "One On Ones": "1v1",
	"Reflexes": "Ref", "Rushing Out (Tendency)": "TRO", "Throwing": "Thr", "Punching": "Pun",
}

// --- Request contract (ticket 04) ---

type WhoToSignPositionRequest struct {
	Position   string   `json:"position"`
	Attributes []string `json:"attributes"`
}

type WhoToSignRequest struct {
	Team              string                     `json:"team"`
	MaxTransferBudget int64                      `json:"maxTransferBudget"`
	MaxWageBudget     int64                      `json:"maxWageBudget"`
	SquadStatus       string                     `json:"squadStatus"`
	Positions         []WhoToSignPositionRequest `json:"positions"`
	PlayerProfile     string                     `json:"playerProfile"`
	FreeText          string                     `json:"freeText"`
}

var whoToSignSquadStatusLabels = map[string]string{
	"star":            "Star player",
	"regular_starter": "Regular starter",
	"rotation":        "Rotation option",
}

// --- find_players tool + squad summary shape (ticket 02, ticket 07) ---

// ScoutPlayerSummary is the trimmed player shape used both for find_players tool
// results and for squad players sent to the model — same serializer for both, so
// the model compares "what I have" against "what's available" on identical fields.
// ScoutPlayerSummary is the shape actually fed to the model (tool results + squad data).
// MBR is deliberately excluded — it's an internal composite rating that isn't capped at
// 100 despite reading like a percentage, and the model was citing raw MBR values in its
// prose in a way that read as a broken/impossible number to users. MBR is still used
// server-side for sorting (via the raw Player struct), just never serialized to the model.
type ScoutPlayerSummary struct {
	UID                 int64          `json:"uid"`
	Name                string         `json:"name"`
	Club                string         `json:"club"`
	Age                 string         `json:"age"`
	ShortPositions      []string       `json:"shortPositions"`
	Overall             int            `json:"overall"`
	CA                  int            `json:"ca"`
	BestRoleOverall     string         `json:"bestRoleOverall"`
	TransferValueAmount int64          `json:"transferValueAmount"`
	WageAmount          int64          `json:"wageAmount"`
	Nationality         string         `json:"nationality"`
	Attributes          map[string]int `json:"attributes"`
}

func buildScoutPlayerSummary(player Player, attributeLongNames []string) ScoutPlayerSummary {
	attrs := make(map[string]int, len(attributeLongNames))
	for _, longName := range attributeLongNames {
		if code, ok := whoToSignAttributeShortCodes[longName]; ok {
			attrs[longName] = player.NumericAttributes[code]
		}
	}
	return ScoutPlayerSummary{
		UID:                 player.UID,
		Name:                player.Name,
		Club:                player.Club,
		Age:                 player.Age,
		ShortPositions:      player.ShortPositions,
		Overall:             player.Overall,
		CA:                  player.CA,
		BestRoleOverall:     player.BestRoleOverall,
		TransferValueAmount: player.TransferValueAmount,
		WageAmount:          player.WageAmount,
		Nationality:         player.Nationality,
		Attributes:          attrs,
	}
}

// whoToSignFindPlayersArgs is the argument shape find_players is declared with (ticket 02).
type whoToSignFindPlayersArgs struct {
	ShortPosition string   `json:"shortPosition"`
	MaxBudget     int64    `json:"maxBudget"`
	MaxSalary     int64    `json:"maxSalary"`
	MinAge        int      `json:"minAge"`
	MaxAge        int      `json:"maxAge"`
	MinOverall    int      `json:"minOverall"`
	Attributes    []string `json:"attributes"`
	FreeText      string   `json:"freeText"`
}

var findPlayersToolSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"shortPosition": map[string]any{
			"type":        "string",
			"description": "The short position code to search for, e.g. \"ST\", \"CB\", \"GK\".",
		},
		"maxBudget":  map[string]any{"type": "number", "description": "Maximum transfer fee, in the dataset's currency units. 0 or omitted means no cap."},
		"maxSalary":  map[string]any{"type": "number", "description": "Maximum weekly wage. 0 or omitted means no cap."},
		"minAge":     map[string]any{"type": "number", "description": "Minimum age. 0 or omitted means no minimum."},
		"maxAge":     map[string]any{"type": "number", "description": "Maximum age. 0 or omitted means no maximum."},
		"minOverall": map[string]any{"type": "number", "description": "Minimum overall rating. 0 or omitted means no minimum."},
		"attributes": map[string]any{
			"type":        "array",
			"items":       map[string]any{"type": "string"},
			"description": "Up to 5 attribute names (long form, e.g. \"Finishing\", \"Pace\") to include in each result's attribute values.",
		},
		"freeText": map[string]any{"type": "string", "description": "Any free-text player-profile preference to keep in mind, for your own reasoning — not applied as a filter."},
	},
	"required": []string{"shortPosition"},
}

// findPlayersForScout is the find_players tool, executed in-process (no HTTP self-hop).
// Filtering/sorting/trimming itself lives in the shared searchPlayers (scout_search.go, ticket
// 02 of the Scout Report map) so Scout Report's find_comparable_players tool doesn't duplicate
// this logic — only the argument translation and attribute-shaped serialization stay here.
func findPlayersForScout(datasetID string, args whoToSignFindPlayersArgs) []ScoutPlayerSummary {
	matches := searchPlayers(datasetID, PlayerSearchCriteria{
		ShortPosition: args.ShortPosition,
		MaxBudget:     args.MaxBudget,
		MaxSalary:     args.MaxSalary,
		MinAge:        args.MinAge,
		MaxAge:        args.MaxAge,
		MinOverall:    args.MinOverall,
	}, whoToSignFindPlayersN)

	summaries := make([]ScoutPlayerSummary, len(matches))
	for i, player := range matches {
		summaries[i] = buildScoutPlayerSummary(player, args.Attributes)
	}
	return summaries
}

// --- Squad lookup (ticket 07) ---

// getSquadForTeam mirrors teamDataHandler's Club-based filter.
func getSquadForTeam(datasetID, team string) ([]Player, bool) {
	players, _, found := GetPlayerData(datasetID)
	if !found {
		return nil, false
	}
	squad := make([]Player, 0)
	for _, player := range players {
		if player.Club == team {
			squad = append(squad, player)
		}
	}
	return squad, true
}

func buildSquadSummaries(squad []Player, req WhoToSignRequest) []ScoutPlayerSummary {
	if len(req.Positions) == 0 {
		summaries := make([]ScoutPlayerSummary, len(squad))
		for i, player := range squad {
			summaries[i] = buildScoutPlayerSummary(player, nil)
		}
		return summaries
	}

	positionAttrs := make(map[string][]string, len(req.Positions))
	requestedPositions := make(map[string]bool, len(req.Positions))
	for _, pos := range req.Positions {
		requestedPositions[strings.ToUpper(pos.Position)] = true
		positionAttrs[strings.ToUpper(pos.Position)] = pos.Attributes
	}

	summaries := make([]ScoutPlayerSummary, 0, len(squad))
	for _, player := range squad {
		var attrs []string
		matched := false
		for _, shortPos := range player.ShortPositions {
			if requestedPositions[strings.ToUpper(shortPos)] {
				matched = true
				attrs = positionAttrs[strings.ToUpper(shortPos)]
				break
			}
		}
		if !matched {
			continue
		}
		summaries = append(summaries, buildScoutPlayerSummary(player, attrs))
	}
	return summaries
}

// whoToSignTableColumns are the fixed leading columns of renderPlayerTable's output, in
// order. Any attribute names present on the given players are appended as extra trailing
// columns (union across all rows, so a mixed-position batch stays one flat table).
var whoToSignTableColumns = []string{"uid", "name", "club", "age", "pos", "ovr", "ca", "role", "value", "wage", "nat"}

// tableFieldSafe strips characters that would corrupt the pipe-delimited table's row/column
// structure. Real FM player/club names don't contain these, but the model is not a trusted
// input source for round-tripped strings, so this is defensive rather than load-bearing.
func tableFieldSafe(s string) string {
	s = strings.ReplaceAll(s, "|", "/")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

// renderPlayerTable encodes player summaries as a compact pipe-delimited table (header row
// first) instead of JSON. Measured on real squad/find_players payloads, this cuts tokens
// roughly in half versus JSON for the same data: JSON repeats every field name (e.g.
// "transferValueAmount", "bestRoleOverall") on every single row, which dominates payload
// size for a uniform list of objects like this one. A table pays for each column name once.
func renderPlayerTable(players []ScoutPlayerSummary) string {
	if len(players) == 0 {
		return "(none)"
	}

	var attrNames []string
	seenAttr := make(map[string]bool)
	for _, p := range players {
		for name := range p.Attributes {
			if !seenAttr[name] {
				seenAttr[name] = true
				attrNames = append(attrNames, name)
			}
		}
	}
	sort.Strings(attrNames)

	cols := make([]string, 0, len(whoToSignTableColumns)+len(attrNames))
	cols = append(cols, whoToSignTableColumns...)
	cols = append(cols, attrNames...)

	var b strings.Builder
	b.WriteString(strings.Join(cols, "|"))
	for _, p := range players {
		row := []string{
			strconv.FormatInt(p.UID, 10),
			tableFieldSafe(p.Name),
			tableFieldSafe(p.Club),
			p.Age,
			strings.Join(p.ShortPositions, ","),
			strconv.Itoa(p.Overall),
			strconv.Itoa(p.CA),
			tableFieldSafe(p.BestRoleOverall),
			strconv.FormatInt(p.TransferValueAmount, 10),
			strconv.FormatInt(p.WageAmount, 10),
			tableFieldSafe(p.Nationality),
		}
		for _, name := range attrNames {
			if v, ok := p.Attributes[name]; ok {
				row = append(row, strconv.Itoa(v))
			} else {
				row = append(row, "")
			}
		}
		b.WriteByte('\n')
		b.WriteString(strings.Join(row, "|"))
	}
	return b.String()
}

// findSquadPlayerByUID looks up a specific squad player by UID — used for the model's own
// choice of who's most relevant to compare mainPick against (ticket: "why did we compare
// against [a reserve]?" — the model already reasons about the whole squad, so trust its
// pick over a disconnected server-side heuristic). Returns nil if the UID isn't an actual
// squad member (defensive against a hallucinated UID) or is the "none" sentinel (0).
func findSquadPlayerByUID(squad []Player, uid int64, attributes []string) *ScoutPlayerSummary {
	if uid == 0 {
		return nil
	}
	for i := range squad {
		if squad[i].UID == uid {
			summary := buildScoutPlayerSummary(squad[i], attributes)
			return &summary
		}
	}
	return nil
}

// findClosestSquadPlayer returns the current squad's best player (by Overall) at the given
// uppercased short position — a defensive fallback for when the model's own
// currentSquadComparisonUid doesn't resolve to an actual squad member. Returns nil if no
// squad player plays that position.
func findClosestSquadPlayer(squad []Player, upperShortPosition string, attributes []string) *ScoutPlayerSummary {
	var best *Player
	for i := range squad {
		player := squad[i]
		hasPosition := false
		for _, pos := range player.ShortPositions {
			if strings.EqualFold(pos, upperShortPosition) {
				hasPosition = true
				break
			}
		}
		if !hasPosition {
			continue
		}
		if best == nil || player.Overall > best.Overall {
			best = &player
		}
	}
	if best == nil {
		return nil
	}
	summary := buildScoutPlayerSummary(*best, attributes)
	return &summary
}

// --- Response schema (ticket 03) ---

// ScoutPlayerRecommendation is the shape used for mainPick/runnersUp in the final
// structured response — ScoutPlayerSummary's fields minus attributes, plus reasoning.
// (Kept as its own flat struct, not embedding ScoutPlayerSummary, so the strict-mode
// JSON schema below and the Go struct tags stay in lockstep.)
type ScoutPlayerRecommendation struct {
	UID                 int64    `json:"uid"`
	Name                string   `json:"name"`
	Club                string   `json:"club"`
	Age                 string   `json:"age"`
	ShortPositions      []string `json:"shortPositions"`
	Overall             int      `json:"overall"`
	CA                  int      `json:"ca"`
	BestRoleOverall     string   `json:"bestRoleOverall"`
	TransferValueAmount int64    `json:"transferValueAmount"`
	WageAmount          int64    `json:"wageAmount"`
	Nationality         string   `json:"nationality"`
	// Grade/SigningScore are the model's own holistic judgement of how good a SIGNING
	// this player would be against the manager's stated requirements — not a restatement
	// of Overall/CA/MBR (ticket: "we should have the LLM choosing their own grade").
	Grade        string   `json:"grade"`
	SigningScore int      `json:"signingScore"`
	Strengths    []string `json:"strengths"`
	WatchOuts    []string `json:"watchOuts"`
}

type WhoToSignPositionRecommendation struct {
	Position          string                      `json:"position"`
	PositionRationale string                      `json:"positionRationale"`
	MainPick          ScoutPlayerRecommendation   `json:"mainPick"`
	RunnersUp         []ScoutPlayerRecommendation `json:"runnersUp"`
	// CurrentSquadComparisonUID is the model's own pick of which current squad player
	// (from the squad list it was given) is most relevant to compare mainPick against —
	// typically whoever mainPick would be competing with for starts. 0 means "no relevant
	// current squad player". The model reasons about the squad as a whole anyway, so this
	// keeps the comparison consistent with whatever it says in positionRationale/watchOuts,
	// rather than a disconnected server-side heuristic picking someone else entirely.
	CurrentSquadComparisonUID int64 `json:"currentSquadComparisonUid"`

	// Populated server-side after the model's structured answer is parsed — not part
	// of the strict-mode schema the model is constrained to, just bookkeeping this
	// backend already has on hand from running find_players and reading the squad.
	PlayersConsidered  []ScoutPlayerSummary `json:"playersConsidered"`
	ClosestSquadPlayer *ScoutPlayerSummary  `json:"closestSquadPlayer,omitempty"`
}

type WhoToSignResponse struct {
	Recommendations []WhoToSignPositionRecommendation `json:"recommendations"`
}

// applyAuthoritativePlayerFields overwrites rec's factual fields (name, club, age,
// positions, overall, ca, role, value, wage, nationality) with the authoritative record
// for rec.UID, checked first against this request's find_players candidate pool and then
// the current squad. If the uid can't be resolved against either (a hallucinated uid, or a
// real player never actually surfaced this request), rec is left exactly as the model
// reported it — this only ever tightens known-good data, never invents a replacement.
// The model's own subjective judgement (grade, signingScore, strengths, watchOuts) is
// never touched here.
func applyAuthoritativePlayerFields(rec *ScoutPlayerRecommendation, squad []Player, candidates map[int64]ScoutPlayerSummary) {
	if summary, ok := candidates[rec.UID]; ok {
		overwriteRecommendationFields(rec, summary)
		return
	}
	for i := range squad {
		if squad[i].UID == rec.UID {
			overwriteRecommendationFields(rec, buildScoutPlayerSummary(squad[i], nil))
			return
		}
	}
}

func overwriteRecommendationFields(rec *ScoutPlayerRecommendation, s ScoutPlayerSummary) {
	rec.Name = s.Name
	rec.Club = s.Club
	rec.Age = s.Age
	rec.ShortPositions = s.ShortPositions
	rec.Overall = s.Overall
	rec.CA = s.CA
	rec.BestRoleOverall = s.BestRoleOverall
	rec.TransferValueAmount = s.TransferValueAmount
	rec.WageAmount = s.WageAmount
	rec.Nationality = s.Nationality
}

var scoutPlayerRecommendationSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"uid":                 map[string]any{"type": "integer"},
		"name":                map[string]any{"type": "string"},
		"club":                map[string]any{"type": "string"},
		"age":                 map[string]any{"type": "string"},
		"shortPositions":      map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
		"overall":             map[string]any{"type": "integer"},
		"ca":                  map[string]any{"type": "integer"},
		"bestRoleOverall":     map[string]any{"type": "string"},
		"transferValueAmount": map[string]any{"type": "integer"},
		"wageAmount":          map[string]any{"type": "integer"},
		"nationality":         map[string]any{"type": "string"},
		"grade": map[string]any{
			"type": "string",
			"enum": []string{"D", "C", "B", "A", "A*"},
			"description": "Your own holistic SCOUT VERDICT grade for this player as a signing against the manager's stated requirements — not a restatement of their Overall/CA rating. " +
				"Anchor it: A/A* = a clear, immediate, undisputed upgrade at good value, no material caveats. B = a good signing with some caveats (age, minor fit gap, slight overpay) but a clear net positive. " +
				"C = a fallback or development pick — the market didn't offer a strong option, or this is a bet on potential rather than a proven upgrade. D = a weak signing or last resort. " +
				"If your own strengths/watchOuts/positionRationale describe this player as a 'fallback', 'value-led development signing', 'not a guaranteed upgrade', or similar hedge, the grade must be C or lower — never grade a hedged pick A or A*.",
		},
		"signingScore": map[string]any{
			"type":        "integer",
			"description": "0-100 holistic SCOUT VERDICT score matching the grade band above: how good a SIGNING this player would be for these specific requirements (position need, budget fit, value, role fit) — not their Overall or CA rating. Must be consistent with the grade (e.g. a C grade should score roughly 45-64, not 80+).",
		},
		"strengths": map[string]any{
			"type":        "array",
			"items":       map[string]any{"type": "string"},
			"description": "Concise bullet points on why this player suits the requirements.",
		},
		"watchOuts": map[string]any{
			"type":        "array",
			"items":       map[string]any{"type": "string"},
			"description": "Concise, honest bullet points on this player's weaknesses or risks as a signing. Never leave this empty — every player has trade-offs.",
		},
	},
	"required": []string{
		"uid", "name", "club", "age", "shortPositions", "overall", "ca",
		"bestRoleOverall", "transferValueAmount", "wageAmount", "nationality",
		"grade", "signingScore", "strengths", "watchOuts",
	},
	"additionalProperties": false,
}

var whoToSignResponseSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"recommendations": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"position":          map[string]any{"type": "string"},
					"positionRationale": map[string]any{"type": "string"},
					"mainPick":          scoutPlayerRecommendationSchema,
					"runnersUp": map[string]any{
						"type":     "array",
						"maxItems": 3,
						"items":    scoutPlayerRecommendationSchema,
					},
					"currentSquadComparisonUid": map[string]any{
						"type":        "integer",
						"description": "The uid (from the current squad list you were given) of whichever squad player is most relevant to compare mainPick against — usually whoever mainPick would be competing with for starts, not necessarily the highest-rated squad player at this position. Use 0 if no current squad player is a relevant comparison.",
					},
				},
				"required":             []string{"position", "positionRationale", "mainPick", "runnersUp", "currentSquadComparisonUid"},
				"additionalProperties": false,
			},
		},
	},
	"required":             []string{"recommendations"},
	"additionalProperties": false,
}

// --- System prompt (ticket 05) ---

func buildWhoToSignSystemPrompt(req WhoToSignRequest, squadSummaries []ScoutPlayerSummary) string {
	positionsText := `none specified — review the squad below and identify 1 to 3 positions most in need of improvement before scouting`
	if len(req.Positions) > 0 {
		var parts []string
		for _, pos := range req.Positions {
			if len(pos.Attributes) > 0 {
				parts = append(parts, pos.Position+" (key attributes: "+strings.Join(pos.Attributes, ", ")+")")
			} else {
				parts = append(parts, pos.Position)
			}
		}
		positionsText = strings.Join(parts, "; ")
	}

	playerProfileText := "no specific preference stated"
	if req.PlayerProfile != "" {
		playerProfileText = req.PlayerProfile
	}
	freeTextText := "none provided"
	if req.FreeText != "" {
		freeTextText = req.FreeText
	}
	squadStatusText := whoToSignSquadStatusLabels[req.SquadStatus]
	if squadStatusText == "" {
		squadStatusText = req.SquadStatus
	}

	squadTable := renderPlayerTable(squadSummaries)

	var b strings.Builder
	b.WriteString("You are an Association Football Manager's head scout, currently employed at ")
	b.WriteString(req.Team)
	b.WriteString(". Your goal is to assist the manager in signing the player(s) most aligned to the manager's requirements below.\n\n")
	b.WriteString("Requirements:\n")
	b.WriteString("- Squad status target for any signing: " + squadStatusText + "\n")
	b.WriteString("- Maximum transfer budget: " + strconv.FormatInt(req.MaxTransferBudget, 10) + "\n")
	b.WriteString("- Maximum wage budget: " + strconv.FormatInt(req.MaxWageBudget, 10) + "\n")
	b.WriteString("- Positions requested: " + positionsText + "\n")
	b.WriteString("- Preferred player profile: " + playerProfileText + "\n")
	b.WriteString("- Manager's own notes: " + freeTextText + "\n\n")
	b.WriteString("Current squad (pipe-delimited table, header row first; pos is a comma-separated list of short position codes):\n")
	b.WriteString(squadTable)
	b.WriteString("\n\n")
	b.WriteString("Instructions:\n")
	b.WriteString("- If positions were specified above, scout for exactly those positions. If none were specified, first identify 1-3 positions from the current squad most in need of strengthening, and scout for those instead.\n")
	b.WriteString("- Use the find_players tool to search the transfer market. It also returns a pipe-delimited table in the same column format as the squad above. You may call it multiple times — once per position, and again with loosened criteria (wider budget/age range, fewer attribute filters) if a search returns few or no suitable options. Only if truly nothing suitable exists after widening the search should you name the closest available player as your pick and say so honestly in your reasoning.\n")
	b.WriteString("- Never state a player's stats, attributes, or attribute-derived claims that didn't come from a find_players result — do not invent numbers.\n")
	b.WriteString("- Use Football Manager terminology and this app's position/attribute naming conventions, not generic football commentary. When you discuss a player's attributes, only ever reference real Football Manager attribute names (e.g. Finishing, Composure, Tackling, Pace, Work Rate) — never reference FIFA-style aggregate ratings such as PAC, SHO, PAS, DRI, DEF, or PHY; those are not attributes and must not appear in your reasoning.\n")
	b.WriteString("- For each position scouted, make a conclusive decision on a single player to sign, and highlight 1 to 3 other good options as runners-up.\n")
	b.WriteString("- For every player you name (main pick and each runner-up), give: a grade (D, C, B, A, or A*) and a signingScore (0-100) that are YOUR OWN holistic judgement of how good a SIGNING they would be for these specific requirements — weighing position need, fit, value for the stated budget, and role suitability. This is not the same as their Overall/CA rating: a modest player who is an excellent, efficient fit for the need should grade higher than a superior player who is a poor fit or bad value, and a hedged pick (a fallback, a development bet, 'not a guaranteed upgrade') must never be graded A or A* — grade it C or lower instead, consistent with what you actually say in your reasoning. Also give strengths (concise bullets on why they suit the requirements, referencing concrete data points) and watchOuts (concise, honest bullets on their weaknesses or risks as a signing) — every player has trade-offs, so watchOuts must never be empty, even for your top pick.\n")
	b.WriteString("- For each position, also set currentSquadComparisonUid to the uid of whichever player already in the squad (from the squad list above) is most relevant to compare your main pick against — usually whoever they'd be competing with for starts. This should be consistent with who you actually discuss in positionRationale/watchOuts, not just whoever happens to have the highest rating. Use 0 if no current squad player is a relevant comparison.\n")
	b.WriteString("- Respond only in the structured format provided.\n")
	return b.String()
}

// --- Request validation + error taxonomy (ticket 10) ---

type whoToSignError struct {
	status  int
	message string
}

func (e *whoToSignError) Error() string { return e.message }

// errToolRoundsExceeded is the sentinel passed to RecordError when a feature's
// tool-calling loop hits its round cap without a final answer — shared by all three
// LLM features (tracing map ticket 02's "tool_rounds_exceeded" reason).
var errToolRoundsExceeded = errors.New("tool-call round cap exceeded without a final answer")

func writeWhoToSignError(w http.ResponseWriter, r *http.Request, err *whoToSignError) {
	setCORSHeaders(w, r)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.message})
}

// --- Handler ---

func whoToSignHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/who-to-sign/"), "/")
	if len(pathParts) == 0 || pathParts[0] == "" {
		http.Error(w, "Dataset ID is missing in the request path", http.StatusBadRequest)
		return
	}
	datasetID := pathParts[0]

	var req WhoToSignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error parsing request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Team == "" || req.MaxTransferBudget <= 0 || req.MaxWageBudget <= 0 || req.SquadStatus == "" {
		writeWhoToSignError(w, r, &whoToSignError{http.StatusBadRequest, "team, maxTransferBudget, maxWageBudget, and squadStatus are all required"})
		return
	}

	apiKey := r.Header.Get("X-OpenAI-Api-Key")
	if apiKey == "" {
		apiKey = whoToSignEnvAPIKey()
	}
	if apiKey == "" {
		writeWhoToSignError(w, r, &whoToSignError{http.StatusBadRequest, "no API key provided — add one in Settings"})
		return
	}
	llmConfig := readLLMRequestConfig(r)

	squad, found := getSquadForTeam(datasetID, req.Team)
	if !found {
		writeWhoToSignError(w, r, &whoToSignError{http.StatusNotFound, "dataset not found"})
		return
	}
	if len(squad) == 0 {
		writeWhoToSignError(w, r, &whoToSignError{http.StatusBadRequest, "no players found for that team — check the team name"})
		return
	}

	logInfo(ctx, "Processing who-to-sign request", "dataset_id", datasetID, "team", req.Team)

	positionsJSON, _ := json.Marshal(req.Positions)
	SetSpanAttributes(ctx,
		attribute.String("fm24.who_to_sign.team", req.Team),
		attribute.String("fm24.who_to_sign.squad_status", req.SquadStatus),
		attribute.String("fm24.who_to_sign.player_profile", req.PlayerProfile),
		attribute.String("fm24.who_to_sign.free_text", req.FreeText),
		attribute.String("fm24.who_to_sign.positions", string(positionsJSON)),
		attribute.Int64("fm24.who_to_sign.max_transfer_budget", req.MaxTransferBudget),
		attribute.Int64("fm24.who_to_sign.max_wage_budget", req.MaxWageBudget),
	)

	squadSummaries := buildSquadSummaries(squad, req)
	systemPrompt := buildWhoToSignSystemPrompt(req, squadSummaries)

	// SSE from here on (Safari fix, see .scratch/llm-refinements/issues/
	// 05-safari-compatibility-investigation.md): all validation above still fails as
	// plain JSON before any bytes are committed to the client.
	setCORSHeaders(w, r)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher, canFlush := w.(http.Flusher)
	sendStatus := func(label string) {
		writeSSEEvent(w, "status", map[string]string{"label": label})
		if canFlush {
			flusher.Flush()
		}
	}

	client, clientErr := newLLMClient(apiKey, llmConfig)
	if clientErr != nil {
		writeSSEEvent(w, "error", map[string]string{"code": strconv.Itoa(http.StatusBadRequest), "message": clientErr.Error()})
		if canFlush {
			flusher.Flush()
		}
		return
	}
	result, considered, candidates, scoutErr := runWhoToSignLoop(ctx, client, datasetID, llmConfig.resolveModel(whoToSignModel), systemPrompt, sendStatus)
	if scoutErr != nil {
		logWarn(ctx, "who-to-sign request failed", "dataset_id", datasetID, "status", scoutErr.status, "error", scoutErr.message)
		writeSSEEvent(w, "error", map[string]string{"code": strconv.Itoa(scoutErr.status), "message": scoutErr.message})
		if canFlush {
			flusher.Flush()
		}
		return
	}

	// Attach the candidate pool and closest current-squad player per position — server
	// bookkeeping the model never needs to produce itself.
	positionAttrs := make(map[string][]string, len(req.Positions))
	for _, pos := range req.Positions {
		positionAttrs[strings.ToUpper(pos.Position)] = pos.Attributes
	}
	playersRecommendedTotal := 0
	for i := range result.Recommendations {
		posKey := strings.ToUpper(result.Recommendations[i].Position)
		result.Recommendations[i].PlayersConsidered = considered[posKey]

		positionRecommendationCount := 1 + len(result.Recommendations[i].RunnersUp)
		playersRecommendedTotal += positionRecommendationCount
		RecordWhoToSignPositionOutcome(ctx, positionRecommendationCount, result.Recommendations[i].MainPick.SigningScore)

		// Overwrite the model's factual fields (name, club, value, wage, ...) with the
		// authoritative record for that uid, wherever one can be resolved — the model's own
		// subjective judgement (grade, signingScore, strengths, watchOuts, positionRationale)
		// is left untouched. This closes off large-number hallucination on mainPick/runnersUp
		// the same way currentSquadComparisonUid already trusts the model's uid choice but
		// not its prose.
		applyAuthoritativePlayerFields(&result.Recommendations[i].MainPick, squad, candidates)
		for j := range result.Recommendations[i].RunnersUp {
			applyAuthoritativePlayerFields(&result.Recommendations[i].RunnersUp[j], squad, candidates)
		}

		// Prefer the model's own pick of who's most relevant to compare against — it
		// already reasoned about the whole squad. Only fall back to the by-Overall
		// heuristic if the model's UID doesn't resolve to an actual squad member.
		closest := findSquadPlayerByUID(squad, result.Recommendations[i].CurrentSquadComparisonUID, positionAttrs[posKey])
		if closest == nil {
			closest = findClosestSquadPlayer(squad, posKey, positionAttrs[posKey])
		}
		result.Recommendations[i].ClosestSquadPlayer = closest
	}
	RecordWhoToSignPlayersRecommendedTotal(ctx, playersRecommendedTotal)
	SetSpanAttributes(ctx, attribute.Int("fm24.who_to_sign.players_recommended_total", playersRecommendedTotal))

	writeSSEEvent(w, "done", result)
	if canFlush {
		flusher.Flush()
	}
}

// whoToSignEnvAPIKey is the optional self-hoster fallback (ticket 06) — never set by
// this app's own deployment configs; the header from the user's own Settings wins
// whenever present.
func whoToSignEnvAPIKey() string {
	return os.Getenv("OPENAI_API_KEY")
}

// runWhoToSignLoop drives the Responses API tool-calling loop: declares find_players,
// executes it in-process on each function_call, and returns once the model produces
// a final structured answer — capped at whoToSignMaxToolRounds rounds (ticket 08),
// with one retry on transient OpenAI errors (ticket 10).
func runWhoToSignLoop(ctx context.Context, client *openai.Client, datasetID, model, systemPrompt string, sendStatus func(string)) (*WhoToSignResponse, map[string][]ScoutPlayerSummary, map[int64]ScoutPlayerSummary, *whoToSignError) {
	// considered accumulates every find_players result across the whole loop, keyed by
	// the uppercased shortPosition searched, deduped by UID and capped at 10 per
	// position — purely server-side bookkeeping for the frontend's "other players
	// considered" table, not something the model needs to know about.
	considered := make(map[string][]ScoutPlayerSummary)
	consideredSeen := make(map[string]map[int64]bool)
	recordConsidered := func(shortPosition string, results []ScoutPlayerSummary) {
		key := strings.ToUpper(shortPosition)
		if consideredSeen[key] == nil {
			consideredSeen[key] = make(map[int64]bool)
		}
		for _, p := range results {
			if len(considered[key]) >= 10 || consideredSeen[key][p.UID] {
				continue
			}
			consideredSeen[key][p.UID] = true
			considered[key] = append(considered[key], p)
		}
	}

	// candidates is every find_players result ever returned this request, uncapped and
	// keyed by UID — the authoritative source used after the model answers to overwrite
	// its factual fields (see applyAuthoritativePlayerFields), so it must hold every player
	// the model could plausibly have picked, not just the capped-at-10 "considered" list
	// kept for the frontend's UI.
	candidates := make(map[int64]ScoutPlayerSummary)
	recordCandidates := func(results []ScoutPlayerSummary) {
		for _, p := range results {
			candidates[p.UID] = p
		}
	}

	tools := []responses.ToolUnionParam{{
		OfFunction: &responses.FunctionToolParam{
			Name:        "find_players",
			Description: openai.String("Search the transfer market for players matching the given position and filters. Returns a pipe-delimited table (header row first): uid|name|club|age|pos|ovr|ca|role|value|wage|nat, plus one column per requested attribute."),
			Parameters:  findPlayersToolSchema,
		},
	}}
	textFormat := responses.ResponseTextConfigParam{
		Format: func() responses.ResponseFormatTextConfigUnionParam {
			schema := responses.ResponseFormatTextJSONSchemaConfigParam{
				Name:   "who_to_sign_response",
				Schema: whoToSignResponseSchema,
				Strict: openai.Bool(true),
			}
			return responses.ResponseFormatTextConfigUnionParam{OfJSONSchema: &schema}
		}(),
	}

	// Tools and Text.Format must be repeated on every request in the loop, including
	// follow-ups chained via PreviousResponseID — they are not carried over automatically,
	// so omitting them on a follow-up leaves the model free to answer in prose instead of
	// the required JSON schema.
	params := responses.ResponseNewParams{
		Model: model,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(systemPrompt),
		},
		Tools: tools,
		Text:  textFormat,
	}

	resp, apiErr := callResponsesWithRetry(ctx, client, params, whoToSignFeature, 0)
	if apiErr != nil {
		return nil, nil, nil, apiErr
	}

	// round counts response round-trips; the response is checked for a final answer
	// both before the first tool call and after every subsequent one, so a legitimate
	// answer on the Nth round is never missed. The model may batch several find_players
	// calls into a single round (parallel tool calls are on by default) — every one of
	// them must get a matching function_call_output in the same follow-up request, or
	// OpenAI rejects the next call with "No tool output found for function call ...".
	for round := 0; ; round++ {
		var toolCalls []responses.ResponseFunctionToolCall
		for i := range resp.Output {
			if resp.Output[i].Type == "function_call" {
				toolCalls = append(toolCalls, resp.Output[i].AsFunctionCall())
			}
		}
		if len(toolCalls) == 0 {
			// No more tool calls — this should be the final structured answer.
			var result WhoToSignResponse
			outputText := resp.OutputText()
			if err := json.Unmarshal([]byte(outputText), &result); err != nil {
				logWarn(ctx, "who-to-sign final response could not be parsed",
					"error", err.Error(),
					"status", resp.Status,
					"incomplete_reason", resp.IncompleteDetails.Reason,
					"output_text_len", len(outputText),
					"output_text", outputText,
				)
				RecordError(ctx, err, "Final response could not be parsed", WithFeature(whoToSignFeature), WithErrorCode("response_unparseable"))
				return nil, nil, nil, &whoToSignError{http.StatusInternalServerError, "the AI scout's response could not be parsed"}
			}
			return &result, considered, candidates, nil
		}
		if round >= whoToSignMaxToolRounds {
			break
		}

		// Flushes bytes to the client every round (research finding, see
		// .scratch/llm-refinements/issues/05-safari-compatibility-investigation.md) —
		// keeps Safari's ~60s idle-fetch timeout from tripping on a call that can
		// legitimately run up to this route's 120s server-side budget.
		sendStatus("Searching players…")

		outputItems := make([]responses.ResponseInputItemUnionParam, len(toolCalls))
		for i, toolCall := range toolCalls {
			var args whoToSignFindPlayersArgs
			_ = json.Unmarshal([]byte(toolCall.Arguments), &args)
			toolResults := findPlayersForScout(datasetID, args)
			RecordToolCall(ctx, "find_players", toolCall.Arguments, len(toolResults))
			recordConsidered(args.ShortPosition, toolResults)
			recordCandidates(toolResults)
			toolResultsTable := renderPlayerTable(toolResults)

			outputItems[i] = responses.ResponseInputItemUnionParam{
				OfFunctionCallOutput: &responses.ResponseInputItemFunctionCallOutputParam{
					CallID: toolCall.CallID,
					Output: responses.ResponseInputItemFunctionCallOutputOutputUnionParam{
						OfString: openai.String(toolResultsTable),
					},
				},
			}
		}

		nextParams := responses.ResponseNewParams{
			Model:              model,
			PreviousResponseID: openai.String(resp.ID),
			Tools:              tools,
			Text:               textFormat,
			Input: responses.ResponseNewParamsInputUnion{
				OfInputItemList: outputItems,
			},
		}
		resp, apiErr = callResponsesWithRetry(ctx, client, nextParams, whoToSignFeature, round+1)
		if apiErr != nil {
			return nil, nil, nil, apiErr
		}
	}

	RecordError(ctx, errToolRoundsExceeded, "Tool-call round cap exceeded", WithFeature(whoToSignFeature), WithErrorCode("tool_rounds_exceeded"))
	return nil, nil, nil, &whoToSignError{
		http.StatusGatewayTimeout,
		"the AI scout couldn't settle on recommendations within the allowed search attempts — try narrowing your request (fewer positions, or a specific position) and try again.",
	}
}

// callResponsesWithRetry makes one Responses API call, retrying once on transient
// failures (429/5xx/network), per ticket 10. feature/round tag the span/metrics/error
// recorded for this call (tracing map ticket 01) — feature is one of whoToSignFeature/
// chatbotFeature/scoutReportFeature, round is the caller's own round counter (0 for the
// pre-loop call, then the loop's current round for each follow-up).
func callResponsesWithRetry(ctx context.Context, client *openai.Client, params responses.ResponseNewParams, feature string, round int) (*responses.Response, *whoToSignError) {
	ctx, span := StartSpan(ctx, "gen_ai.responses.create")
	defer span.End()
	span.SetAttributes(
		attribute.String("gen_ai.system", "openai"),
		attribute.String("gen_ai.request.model", params.Model),
		attribute.String("fm24.llm.feature", feature),
		attribute.Int("fm24.llm.round", round),
	)

	start := time.Now()
	resp, err := client.Responses.New(ctx, params)
	if err == nil {
		recordLLMCallSuccess(ctx, span, feature, round, resp, time.Since(start))
		return resp, nil
	}

	// A client-side timeout (localLLMTimeout, or the request's own context deadline)
	// means the model call itself is stuck or too slow -- retrying immediately won't
	// make it faster, and burning more time here risks losing the race against
	// main.go's outer 120s RequestTimeoutMiddlewareFunc, whose own timeout path is
	// broken for these SSE routes (see localLLMTimeout's comment). Fail once, cleanly,
	// well inside that budget instead.
	if errors.Is(err, context.DeadlineExceeded) {
		logWarn(ctx, "who-to-sign OpenAI call timed out", "elapsed", time.Since(start).String())
		RecordError(ctx, err, "LLM call timed out", WithFeature(feature), WithErrorCode("llm_timeout"))
		return nil, &whoToSignError{http.StatusGatewayTimeout, "the AI assistant took too long to respond — the model may be overloaded or too slow for this request, try again or try a different model"}
	}

	var apiErr *openai.Error
	if errors.As(err, &apiErr) {
		logWarn(ctx, "who-to-sign OpenAI call failed", "status_code", apiErr.StatusCode, "code", apiErr.Code, "type", apiErr.Type, "message", apiErr.Message)
		if apiErr.StatusCode == http.StatusUnauthorized {
			RecordError(ctx, err, "OpenAI rejected the API key", WithFeature(feature), WithErrorCode("openai_auth_rejected"))
			return nil, &whoToSignError{http.StatusBadGateway, "OpenAI rejected the provided API key — check it's valid in Settings"}
		}
		if apiErr.StatusCode != http.StatusTooManyRequests && apiErr.StatusCode < 500 {
			RecordError(ctx, err, "OpenAI rejected the request", WithFeature(feature), WithErrorCode("openai_request_rejected"))
			return nil, &whoToSignError{http.StatusBadGateway, "OpenAI rejected the request: " + apiErr.Message}
		}
	} else {
		logWarn(ctx, "who-to-sign OpenAI call failed (non-API error)", "error", err.Error())
	}

	// Transient (429/5xx/network) — one retry after a short backoff.
	time.Sleep(500 * time.Millisecond)
	start = time.Now()
	resp, err = client.Responses.New(ctx, params)
	if err != nil {
		var retryAPIErr *openai.Error
		if errors.As(err, &retryAPIErr) {
			logWarn(ctx, "who-to-sign OpenAI retry also failed", "status_code", retryAPIErr.StatusCode, "code", retryAPIErr.Code, "type", retryAPIErr.Type, "message", retryAPIErr.Message)
			RecordError(ctx, err, "OpenAI rejected the retried request", WithFeature(feature), WithErrorCode("openai_unavailable"))
			return nil, &whoToSignError{http.StatusBadGateway, "OpenAI rejected the request: " + retryAPIErr.Message}
		}
		logWarn(ctx, "who-to-sign OpenAI retry also failed (non-API error)", "error", err.Error())
		RecordError(ctx, err, "OpenAI call failed after retry", WithFeature(feature), WithErrorCode("openai_unavailable"))
		return nil, &whoToSignError{http.StatusBadGateway, "the AI scouting service is temporarily unavailable, try again shortly"}
	}
	recordLLMCallSuccess(ctx, span, feature, round, resp, time.Since(start))
	return resp, nil
}

// recordLLMCallSuccess records the GenAI span attributes and the per-call token/duration
// metrics for one successful Responses API call (tracing map ticket 01).
func recordLLMCallSuccess(ctx context.Context, span trace.Span, feature string, round int, resp *responses.Response, duration time.Duration) {
	span.SetAttributes(
		attribute.String("gen_ai.response.model", string(resp.Model)),
		attribute.Int64("gen_ai.usage.input_tokens", resp.Usage.InputTokens),
		attribute.Int64("gen_ai.usage.output_tokens", resp.Usage.OutputTokens),
	)
	span.SetStatus(codes.Ok, "LLM call succeeded")
	RecordLLMCall(ctx, feature, round, resp.Usage.InputTokens, resp.Usage.OutputTokens, duration)
}
