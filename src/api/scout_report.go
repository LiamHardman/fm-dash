package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

// AI Scout Report — implements the spec charted in .scratch/scout-report/ (Wayfinder map
// "AI Scout Report"). Reuses Who to Sign's infrastructure throughout: same model, same
// Responses API tool-calling shape, same BYOK header, same error taxonomy, same
// strict-Structured-Outputs constraints — see who_to_sign.go and its own map for the
// underlying research (ticket 01 there).

const (
	scoutReportModel         = whoToSignModel
	scoutReportMaxToolRounds = 5
	scoutReportFindPlayersN  = 10

	// scoutReportFeature tags this feature's spans/metrics/errors (tracing map ticket 01).
	scoutReportFeature = "scout_report"
)

// scoutReportGradeOrdinal maps scoutReportGradeEnum's letter grades onto the tracing
// map's locked D=1..A+=7 ordinal scale, so grades can be averaged (tracing map ticket 05).
func scoutReportGradeOrdinal(grade string) int {
	for i, g := range scoutReportGradeEnum {
		if g == grade {
			return len(scoutReportGradeEnum) - i // A+=7 ... D=1
		}
	}
	return 0 // unrecognized — shouldn't happen, strict schema constrains the model to the enum
}

// --- Request contract (ticket 04) ---

// PlayerUID is a string on the wire (matches playerPercentilesHandler's
// PlayerPercentilesRequest convention — frontend player objects carry uid as a string),
// parsed to int64 in the handler.
type ScoutReportRequest struct {
	PlayerUID string `json:"playerUid"`
	Position  string `json:"position"`
}

// --- find_comparable_players tool (ticket 02) ---

type findComparablePlayersArgs struct {
	ShortPosition string `json:"shortPosition"`
	MinCA         int    `json:"minCA"`
	MaxCA         int    `json:"maxCA"`
	MinAge        int    `json:"minAge"`
	MaxAge        int    `json:"maxAge"`
	MaxBudget     int64  `json:"maxBudget"`
}

var findComparablePlayersToolSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"shortPosition": map[string]any{
			"type":        "string",
			"description": "The short position code to search for, e.g. \"ST\", \"CB\", \"WBL\".",
		},
		"minCA":     map[string]any{"type": "number", "description": "Minimum internal ability estimate. Use this to set a sensible lower bound around the subject player's own ability. 0 or omitted means no minimum."},
		"maxCA":     map[string]any{"type": "number", "description": "Maximum internal ability estimate. Use this to set a sensible upper bound around the subject player's own ability. 0 or omitted means no maximum."},
		"minAge":    map[string]any{"type": "number", "description": "Minimum age. 0 or omitted means no minimum."},
		"maxAge":    map[string]any{"type": "number", "description": "Maximum age. 0 or omitted means no maximum."},
		"maxBudget": map[string]any{"type": "number", "description": "Maximum transfer fee, in the dataset's currency units. 0 or omitted means no cap."},
	},
	"required": []string{"shortPosition"},
}

// findComparablePlayersForScout is the find_comparable_players tool, executed in-process.
// Built on the same shared searchPlayers as find_players (who_to_sign.go), with its own
// narrower argument shape — no attributes/freeText params, since Scout Report's pros/cons
// are grounded in the subject's own percentiles, not per-candidate attribute lookups.
func findComparablePlayersForScout(datasetID string, args findComparablePlayersArgs, excludeUID int64) []Player {
	return searchPlayers(datasetID, PlayerSearchCriteria{
		ShortPosition: args.ShortPosition,
		MinCA:         args.MinCA,
		MaxCA:         args.MaxCA,
		MinAge:        args.MinAge,
		MaxAge:        args.MaxAge,
		MaxBudget:     args.MaxBudget,
		ExcludeUID:    excludeUID,
	}, scoutReportFindPlayersN)
}

// --- Response schema (ticket 03) ---

var scoutReportGradeEnum = []string{"A+", "A", "B+", "B", "C+", "C", "D"}

// llmScoutReportOutput is the raw strict-mode structured output shape the model is
// constrained to. Comparable players are {uid, grade, rationale} only — the backend
// resolves every factual field by uid afterward (mirrors Who to Sign's
// applyAuthoritativePlayerFields anti-hallucination fix) and computes stars itself
// (scout_report_stars.go), never asked of the LLM.
type llmScoutReportOutput struct {
	SubjectGrade          string                         `json:"subjectGrade"`
	SubjectGradeRationale string                         `json:"subjectGradeRationale"`
	Pros                  []string                       `json:"pros"`
	Cons                  []string                       `json:"cons"`
	ComparablePlayers     []llmScoutReportComparablePick `json:"comparablePlayers"`
}

type llmScoutReportComparablePick struct {
	UID       int64  `json:"uid"`
	Grade     string `json:"grade"`
	Rationale string `json:"rationale"`
}

// ScoutReportComparable is a fully-resolved comparable player for the frontend — CA is
// deliberately never included here (it's an internal value driving the stars/grade
// computation, not a fact to show the manager directly, per this map's Notes).
type ScoutReportComparable struct {
	UID                 int64    `json:"uid"`
	Name                string   `json:"name"`
	Club                string   `json:"club"`
	Age                 string   `json:"age"`
	ShortPositions      []string `json:"shortPositions"`
	Overall             int      `json:"overall"`
	BestRoleOverall     string   `json:"bestRoleOverall"`
	TransferValueAmount int64    `json:"transferValueAmount"`
	WageAmount          int64    `json:"wageAmount"`
	Nationality         string   `json:"nationality"`
	Grade               string   `json:"grade"`
	Rationale           string   `json:"rationale"`
	SquadStars          float64  `json:"squadStars"`
	DivisionStars       float64  `json:"divisionStars"`
}

type ScoutReportResponse struct {
	SubjectGrade          string                  `json:"subjectGrade"`
	SubjectGradeRationale string                  `json:"subjectGradeRationale"`
	Pros                  []string                `json:"pros"`
	Cons                  []string                `json:"cons"`
	ComparablePlayers     []ScoutReportComparable `json:"comparablePlayers"`
}

var scoutReportResponseSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"subjectGrade": map[string]any{
			"type": "string",
			"enum": scoutReportGradeEnum,
		},
		"subjectGradeRationale": map[string]any{"type": "string"},
		"pros": map[string]any{
			"type":        "array",
			"items":       map[string]any{"type": "string"},
			"description": "2-5 concise, specific pros grounded in concrete data (FM attributes, percentiles, value, wage).",
		},
		"cons": map[string]any{
			"type":        "array",
			"items":       map[string]any{"type": "string"},
			"description": "2-5 concise, specific cons grounded in concrete data (FM attributes, percentiles, value, wage).",
		},
		"comparablePlayers": map[string]any{
			"type":     "array",
			"maxItems": 5,
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"uid":       map[string]any{"type": "integer", "description": "The uid of a player returned by find_comparable_players."},
					"grade":     map[string]any{"type": "string", "enum": scoutReportGradeEnum},
					"rationale": map[string]any{"type": "string"},
				},
				"required":             []string{"uid", "grade", "rationale"},
				"additionalProperties": false,
			},
			"description": "Up to 5 comparable players you found, ordered best grade first.",
		},
	},
	"required":             []string{"subjectGrade", "subjectGradeRationale", "pros", "cons", "comparablePlayers"},
	"additionalProperties": false,
}

// --- Grading rubric + system prompt (ticket 03) ---

const scoutReportGradingRubric = `Grade using this rubric, weighing all four factors together rather than any single one:
- Price: transfer value and wage relative to the comparable players found via find_comparable_players. Meaningfully cheaper for similar or better ability grades up. Meaningfully pricier for similar or worse ability grades down.
- Ability: internal ability estimate relative to the strongest players in the same position on the manager's own squad (their current options), not just the wider pool.
- Age: younger players with room to develop grade slightly up versus an otherwise-equal older player; this is a minor factor, not a dominant one.
- Wage demand: likely wage relative to comparable players' wages — grade down if a player's profile suggests wage demands disproportionate to their ability/output.

A+/A: clearly a strong signing/asset — better value and/or ability than nearly all comparables.
B+/B: solid, sensible option — roughly in line with comparables on price and ability.
C+/C: usable but with a real drawback (overpriced, ability below comparables, or a wage-demand concern) — say which in the rationale.
D: hard to justify on the evidence found — significantly overpriced, outclassed by comparables, or both.

For the subject player's own subjectGrade (a player already at their club, not necessarily on the market): apply the same rubric using their actual current wage/value in place of "price relative to comparables", and weigh ability-vs-squad and age normally.`

func buildScoutReportSystemPrompt(managedClub, managedDivision string, subject ScoutPlayerSummary, position string, percentiles map[string]map[string]float64) string {
	var b strings.Builder
	b.WriteString("You are an Association Football Manager's head scout, currently employed at ")
	b.WriteString(managedClub)
	b.WriteString(" (")
	b.WriteString(managedDivision)
	b.WriteString("). You are writing a scout report on ")
	b.WriteString(subject.Name)
	b.WriteString(" (")
	b.WriteString(position)
	b.WriteString(").\n\n")

	b.WriteString("Player being assessed:\n")
	b.WriteString(renderPlayerTable([]ScoutPlayerSummary{subject}))
	b.WriteString("\n\n")

	b.WriteString("This player's percentile standing within their division (100 = best in division at this stat, 0 = worst):\n")
	b.WriteString(renderPercentileTable(percentiles))
	b.WriteString("\n\n")

	b.WriteString(scoutReportGradingRubric)
	b.WriteString("\n\n")

	b.WriteString("Instructions:\n")
	b.WriteString("- Use the find_comparable_players tool to search for players of similar age, ability, and position to ")
	b.WriteString(subject.Name)
	b.WriteString(". You may call it multiple times — narrower or wider ability/age windows — if the first search returns too few or too many options. Cap yourself at a sensible search; you do not need to exhaust every possible window.\n")
	b.WriteString("- From the results, select up to 5 to spotlight in comparablePlayers, each with your own grade and a short, specific rationale, ordered best grade first.\n")
	b.WriteString("- Assign " + subject.Name + " their own subjectGrade using the same rubric.\n")
	b.WriteString("- Write 2-5 pros and 2-5 cons for " + subject.Name + ", grounded in concrete numbers from the data above (attributes, percentiles, value, wage) — never invent a stat, attribute, or percentile that isn't given to you or returned by the tool.\n")
	b.WriteString("- Use Football Manager terminology and this app's position/attribute naming conventions, not generic football commentary. Never reference FIFA-style aggregate ratings such as PAC, SHO, PAS, DRI, DEF, or PHY.\n")
	b.WriteString("- Never state or imply a numeric \"CA\"/ability-estimate value — it is an internal figure, not something managers see in-game. Reason about ability qualitatively (attributes, percentiles, role suitability) instead.\n")
	b.WriteString("- Respond only in the structured format provided.\n")
	return b.String()
}

// renderPercentileTable encodes a player's percentile map (group -> stat -> percentile) as
// a compact pipe-delimited table, same token-minimization approach as renderPlayerTable.
func renderPercentileTable(percentiles map[string]map[string]float64) string {
	if len(percentiles) == 0 {
		return "(no percentile data available)"
	}
	var b strings.Builder
	b.WriteString("group|stat|percentile")
	for group, stats := range percentiles {
		for stat, value := range stats {
			b.WriteByte('\n')
			b.WriteString(tableFieldSafe(group))
			b.WriteByte('|')
			b.WriteString(tableFieldSafe(stat))
			b.WriteByte('|')
			b.WriteString(strconv.Itoa(int(value)))
		}
	}
	return b.String()
}

// --- Percentile lookup (ticket 03) ---

// getSubjectDivisionPercentiles computes the subject player's percentile standing within
// their own division (DivisionFilterSame — the same comparison pool the division-scope
// star rating uses, per this map's decision that both should tell one consistent story).
// Mirrors playerPercentilesHandler's filter-then-recalculate pattern (handlers.go).
func getSubjectDivisionPercentiles(players []Player, subject Player) map[string]map[string]float64 {
	filtered := make([]Player, 0, len(players))
	found := false
	for _, p := range players {
		if isPlayerInTargetDivision(&p, DivisionFilterSame, subject.Division) {
			filtered = append(filtered, p)
			if p.UID == subject.UID {
				found = true
			}
		}
	}
	if !found {
		filtered = append(filtered, subject)
	}

	CalculatePlayerPerformancePercentilesWithDivisionFilter(filtered, DivisionFilterAll, "")

	for _, p := range filtered {
		if p.UID == subject.UID {
			return p.PerformancePercentiles
		}
	}
	return nil
}

// --- Error taxonomy (reused from Who to Sign as-is, per this map's Decisions) ---

func writeScoutReportError(w http.ResponseWriter, r *http.Request, err *whoToSignError) {
	writeWhoToSignError(w, r, err)
}

// --- Handler ---

// scoutReportHandler dispatches GET (fetch one persisted report), DELETE (remove one),
// and POST (generate + persist) on /api/scout-report/{datasetId} — Scout Report v2 map
// ticket 01's extension of what was previously a POST-only endpoint. GET/DELETE are
// handled by scout_report_book.go; POST keeps its own SSE streaming shape here.
func scoutReportHandler(w http.ResponseWriter, r *http.Request) {
	datasetID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/scout-report/"), "/")
	if datasetID == "" {
		http.Error(w, "Dataset ID is missing in the request path", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		scoutReportGetHandler(w, r, datasetID)
	case http.MethodDelete:
		scoutReportDeleteHandler(w, r, datasetID)
	case http.MethodPost:
		scoutReportPostHandler(w, r, datasetID)
	default:
		http.Error(w, "Only GET, POST, and DELETE methods are allowed", http.StatusMethodNotAllowed)
	}
}

func scoutReportPostHandler(w http.ResponseWriter, r *http.Request, datasetID string) {
	ctx := r.Context()

	var req ScoutReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error parsing request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.PlayerUID == "" || req.Position == "" {
		writeScoutReportError(w, r, &whoToSignError{http.StatusBadRequest, "playerUid and position are both required"})
		return
	}
	playerUID, err := strconv.ParseInt(req.PlayerUID, 10, 64)
	if err != nil {
		writeScoutReportError(w, r, &whoToSignError{http.StatusBadRequest, "invalid playerUid"})
		return
	}

	apiKey := r.Header.Get("X-OpenAI-Api-Key")
	if apiKey == "" {
		apiKey = whoToSignEnvAPIKey()
	}
	if apiKey == "" {
		writeScoutReportError(w, r, &whoToSignError{http.StatusBadRequest, "no API key provided — add one in Settings"})
		return
	}
	llmConfig := readLLMRequestConfig(r)

	logInfo(ctx, "Processing scout-report request", "dataset_id", datasetID, "player_uid", req.PlayerUID, "position", req.Position)

	// SSE from here on (Safari fix, see .scratch/llm-refinements/issues/
	// 05-safari-compatibility-investigation.md). generateScoutReport does its own
	// dataset/player/managed-team validation and reports failure via the SSE "error"
	// event below — this endpoint no longer pre-validates before opening the stream
	// (Scout Report v2 map ticket 03: the same generateScoutReport is called from
	// chat's tool dispatch, which has no separate pre-flight phase, so validation lives
	// in one place for both callers rather than being duplicated).
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

	result, scoutErr := generateScoutReport(ctx, client, datasetID, llmConfig.resolveModel(scoutReportModel), playerUID, req.Position, sendStatus)
	if scoutErr != nil {
		logWarn(ctx, "scout-report request failed", "dataset_id", datasetID, "status", scoutErr.status, "error", scoutErr.message)
		writeSSEEvent(w, "error", map[string]string{"code": strconv.Itoa(scoutErr.status), "message": scoutErr.message})
		if canFlush {
			flusher.Flush()
		}
		return
	}

	writeSSEEvent(w, "done", result)
	if canFlush {
		flusher.Flush()
	}
}

// generateScoutReport runs the full pipeline — resolve subject, load managed team,
// build percentiles/prompt, drive the tool-calling loop, resolve comparable players,
// compute stars, persist — shared by scoutReportPostHandler (HTTP) and the chatbot's
// generate_scout_report tool (chatbot.go), so generation and persistence logic is never
// duplicated between the two entry points (Scout Report v2 map tickets 01/03).
func generateScoutReport(ctx context.Context, client *openai.Client, datasetID, model string, playerUID int64, position string, sendStatus func(string)) (*ScoutReportWithMeta, *whoToSignError) {
	players, _, found := GetPlayerData(datasetID)
	if !found {
		return nil, &whoToSignError{http.StatusNotFound, "dataset not found"}
	}
	players = RecalculateAllPlayersRatings(players)

	var subject *Player
	for i := range players {
		if players[i].UID == playerUID {
			subject = &players[i]
			break
		}
	}
	if subject == nil {
		return nil, &whoToSignError{http.StatusNotFound, "player not found in dataset"}
	}

	// Position defaults to the subject's own primary short position when omitted —
	// mirrors ScoutReportTab.vue's own defaultPosition(), so a chat-triggered report
	// (ticket 03, whose tool contract makes shortPosition optional) and a
	// dialog-triggered one agree by default.
	if position == "" {
		if len(subject.ShortPositions) == 0 {
			return nil, &whoToSignError{http.StatusBadRequest, "player has no known position"}
		}
		position = subject.ShortPositions[0]
	}

	managedTeam, err := loadManagedTeam(datasetID)
	if err != nil {
		return nil, &whoToSignError{http.StatusBadRequest, "set your managed team before generating a Scout Report"}
	}

	subjectSummary := buildScoutPlayerSummary(*subject, nil)
	percentiles := getSubjectDivisionPercentiles(players, *subject)
	systemPrompt := buildScoutReportSystemPrompt(managedTeam.Club, managedTeam.Division, subjectSummary, position, percentiles)

	logInfo(ctx, "Generating scout report", "dataset_id", datasetID, "player_uid", playerUID, "position", position)

	llmResult, candidates, scoutErr := runScoutReportLoop(ctx, client, datasetID, model, systemPrompt, subject.UID, sendStatus)
	if scoutErr != nil {
		return nil, scoutErr
	}

	RecordScoutReportSubjectGrade(ctx, scoutReportGradeOrdinal(llmResult.SubjectGrade))

	result := ScoutReportResponse{
		SubjectGrade:          llmResult.SubjectGrade,
		SubjectGradeRationale: llmResult.SubjectGradeRationale,
		Pros:                  llmResult.Pros,
		Cons:                  llmResult.Cons,
	}
	for _, pick := range llmResult.ComparablePlayers {
		candidate, ok := candidates[pick.UID]
		if !ok {
			// Hallucinated or never-surfaced uid — drop it rather than show fabricated data.
			continue
		}
		squadStars, divisionStars := scoutReportStars(players, candidate, position, managedTeam.Club)
		RecordScoutReportComparable(ctx, scoutReportGradeOrdinal(pick.Grade), squadStars, divisionStars)
		result.ComparablePlayers = append(result.ComparablePlayers, ScoutReportComparable{
			UID:                 candidate.UID,
			Name:                candidate.Name,
			Club:                candidate.Club,
			Age:                 candidate.Age,
			ShortPositions:      candidate.ShortPositions,
			Overall:             candidate.Overall,
			BestRoleOverall:     candidate.BestRoleOverall,
			TransferValueAmount: candidate.TransferValueAmount,
			WageAmount:          candidate.WageAmount,
			Nationality:         candidate.Nationality,
			Grade:               pick.Grade,
			Rationale:           pick.Rationale,
			SquadStars:          squadStars,
			DivisionStars:       divisionStars,
		})
	}

	subjectSquadStars, subjectDivisionStars := scoutReportStars(players, *subject, position, managedTeam.Club)
	generatedAt := time.Now()
	if err := saveScoutReport(datasetID, subject.UID, position, subject.Name, subject.Club, result, subjectSquadStars, subjectDivisionStars, generatedAt); err != nil {
		// Persistence is best-effort: a manager's freshly-generated report is still
		// worth returning even if the Scouting Book write failed — they just won't see
		// it there until a future successful save.
		logWarn(ctx, "scout-report: failed to persist report", "dataset_id", datasetID, "player_uid", playerUID, "error", err.Error())
	}

	return &ScoutReportWithMeta{ScoutReportResponse: result, GeneratedAt: generatedAt}, nil
}

// runScoutReportLoop drives the Responses API tool-calling loop for find_comparable_players,
// mirroring runWhoToSignLoop's shape exactly (who_to_sign.go) — same retry/error/round-cap
// handling, reused as-is per this map's Decisions.
func runScoutReportLoop(ctx context.Context, client *openai.Client, datasetID, model, systemPrompt string, subjectUID int64, sendStatus func(string)) (*llmScoutReportOutput, map[int64]Player, *whoToSignError) {
	candidates := make(map[int64]Player)
	recordCandidates := func(results []Player) {
		for _, p := range results {
			candidates[p.UID] = p
		}
	}

	tools := []responses.ToolUnionParam{{
		OfFunction: &responses.FunctionToolParam{
			Name:        "find_comparable_players",
			Description: openai.String("Search for players of similar age, ability, and position to the subject player. Returns a pipe-delimited table (header row first): uid|name|club|age|pos|ovr|ca|role|value|wage|nat."),
			Parameters:  findComparablePlayersToolSchema,
		},
	}}
	textFormat := responses.ResponseTextConfigParam{
		Format: func() responses.ResponseFormatTextConfigUnionParam {
			schema := responses.ResponseFormatTextJSONSchemaConfigParam{
				Name:   "scout_report_response",
				Schema: scoutReportResponseSchema,
				Strict: openai.Bool(true),
			}
			return responses.ResponseFormatTextConfigUnionParam{OfJSONSchema: &schema}
		}(),
	}

	params := responses.ResponseNewParams{
		Model: model,
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(systemPrompt),
		},
		Tools: tools,
		Text:  textFormat,
	}

	resp, apiErr := callResponsesWithRetry(ctx, client, params, scoutReportFeature, 0)
	if apiErr != nil {
		return nil, nil, apiErr
	}

	for round := 0; ; round++ {
		var toolCalls []responses.ResponseFunctionToolCall
		for i := range resp.Output {
			if resp.Output[i].Type == "function_call" {
				toolCalls = append(toolCalls, resp.Output[i].AsFunctionCall())
			}
		}
		if len(toolCalls) == 0 {
			var result llmScoutReportOutput
			outputText := resp.OutputText()
			if err := json.Unmarshal([]byte(outputText), &result); err != nil {
				logWarn(ctx, "scout-report final response could not be parsed",
					"error", err.Error(),
					"status", resp.Status,
					"incomplete_reason", resp.IncompleteDetails.Reason,
					"output_text_len", len(outputText),
					"output_text", outputText,
				)
				RecordError(ctx, err, "Final response could not be parsed", WithFeature(scoutReportFeature), WithErrorCode("response_unparseable"))
				return nil, nil, &whoToSignError{http.StatusInternalServerError, "the AI scout's response could not be parsed"}
			}
			return &result, candidates, nil
		}
		if round >= scoutReportMaxToolRounds {
			break
		}

		// Flushes bytes to the client every round (research finding, see
		// .scratch/llm-refinements/issues/05-safari-compatibility-investigation.md) —
		// keeps Safari's ~60s idle-fetch timeout from tripping on a call that can
		// legitimately run up to this route's 120s server-side budget.
		sendStatus("Finding comparable players…")

		outputItems := make([]responses.ResponseInputItemUnionParam, len(toolCalls))
		for i, toolCall := range toolCalls {
			var args findComparablePlayersArgs
			_ = json.Unmarshal([]byte(toolCall.Arguments), &args)
			toolResults := findComparablePlayersForScout(datasetID, args, subjectUID)
			RecordToolCall(ctx, "find_comparable_players", toolCall.Arguments, len(toolResults))
			RecordScoutReportSearchResults(ctx, len(toolResults))
			recordCandidates(toolResults)

			summaries := make([]ScoutPlayerSummary, len(toolResults))
			for i, p := range toolResults {
				summaries[i] = buildScoutPlayerSummary(p, nil)
			}
			toolResultsTable := renderPlayerTable(summaries)

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
		resp, apiErr = callResponsesWithRetry(ctx, client, nextParams, scoutReportFeature, round+1)
		if apiErr != nil {
			return nil, nil, apiErr
		}
	}

	RecordError(ctx, errToolRoundsExceeded, "Tool-call round cap exceeded", WithFeature(scoutReportFeature), WithErrorCode("tool_rounds_exceeded"))
	return nil, nil, &whoToSignError{
		http.StatusGatewayTimeout,
		"the AI scout couldn't settle on a report within the allowed search attempts — try again shortly.",
	}
}
