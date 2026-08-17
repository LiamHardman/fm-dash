package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

// FM-Dash Chatbot — dataset-scoped, multi-turn LLM scouting assistant. Implements the
// spec charted in .scratch/chatbot/ (Wayfinder map "FM-Dash Chatbot").
//
// Unlike Who to Sign / AI Scout Report (single form submission -> one LLM turn -> one
// structured response), this is a genuine multi-turn chat: stateless per HTTP request,
// full history resent every turn (ticket 01), but history is compressed to plain
// {role, text} turns — raw tool-call scaffolding never persists across turns, so every
// turn's tool-calling loop starts fresh. Delivery is Server-Sent Events (ticket 06):
// staged per-tool-call status labels, then one final complete payload.
//
// Per the map's "fully independent" decision, the four tools below are new,
// self-contained implementations — not calls into bargainHunterHandler/
// WonderkidsDialog/WhoToSignDialog's find_players, even where the domain overlaps.
// Shared infra (BYOK key handling, the openai-go SDK, the retry helper, the managed-team
// storage lookup) is reused as plumbing, per the map's Notes.

const (
	chatbotModel                = "gpt-5.6-luna"
	chatbotMaxToolRoundsPerTurn = 5
	chatbotMaxHistoryTurns      = 50
	chatbotSearchDefaultLimit   = 15
	chatbotSearchMaxLimit       = 100
)

// --- Request/response contract (ticket 01, 06) ---

type ChatHistoryTurn struct {
	Role string `json:"role"` // "user" | "assistant"
	Text string `json:"text"`
}

type ChatRequest struct {
	History []ChatHistoryTurn `json:"history"`
}

type ChatReferencedPlayer struct {
	UID  int64  `json:"uid"`
	Name string `json:"name"`
}

type ChatChart struct {
	Template string         `json:"template"`
	Data     map[string]any `json:"data"`
}

// ChatDoneEvent is the final SSE "done" event payload (ticket 06).
type ChatDoneEvent struct {
	Text              string                 `json:"text"`
	ReferencedPlayers []ChatReferencedPlayer `json:"referencedPlayers"`
	Chart             *ChatChart             `json:"chart,omitempty"`
}

// chatModelResponse is the model's own strict-mode structured output — deliberately
// narrower than ChatDoneEvent. Chart is never part of what the model produces (ticket
// 02/04): it's a callable tool, and the backend attaches whatever render_chart last
// computed when building the final SSE event, so the model never restates chart data.
type chatModelResponse struct {
	Text              string                 `json:"text"`
	ReferencedPlayers []ChatReferencedPlayer `json:"referencedPlayers"`
}

var chatResponseSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"text": map[string]any{
			"type":        "string",
			"description": "Your reply, written as natural chat prose. Wherever you mention a specific player by name, write the token [[player:UID]] in its place (e.g. \"[[player:104829]] would be a strong pick\") instead of the name.",
		},
		"referencedPlayers": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"uid":  map[string]any{"type": "integer"},
					"name": map[string]any{"type": "string"},
				},
				"required":             []string{"uid", "name"},
				"additionalProperties": false,
			},
			"description": "Every player cited via a [[player:UID]] token above, in the order first mentioned. Only cite players that came from a tool result this conversation.",
		},
	},
	"required":             []string{"text", "referencedPlayers"},
	"additionalProperties": false,
}

// --- ChatPlayerSummary (ticket 09) — same field list as Who to Sign's ScoutPlayerSummary,
// separately declared so the two features stay type-level independent. ---

type ChatPlayerSummary struct {
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

// buildChatPlayerSummary reuses whoToSignAttributeShortCodes (who_to_sign.go) — a static
// long-name -> short-code dictionary, pure data plumbing rather than scouting business
// logic, so reusing it doesn't conflict with the "fully independent" decision.
func buildChatPlayerSummary(player Player, attributeLongNames []string) ChatPlayerSummary {
	attrs := make(map[string]int, len(attributeLongNames))
	for _, longName := range attributeLongNames {
		if code, ok := whoToSignAttributeShortCodes[longName]; ok {
			attrs[longName] = player.NumericAttributes[code]
		}
	}
	return ChatPlayerSummary{
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

var chatPlayerTableColumns = []string{"uid", "name", "club", "age", "pos", "ovr", "ca", "role", "value", "wage", "nat"}

func chatTableFieldSafe(s string) string {
	s = strings.ReplaceAll(s, "|", "/")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

// renderChatPlayerTable mirrors who_to_sign.go's renderPlayerTable — a pipe-delimited
// table cuts tokens roughly in half versus JSON for a uniform list of player objects.
func renderChatPlayerTable(players []ChatPlayerSummary) string {
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

	cols := make([]string, 0, len(chatPlayerTableColumns)+len(attrNames))
	cols = append(cols, chatPlayerTableColumns...)
	cols = append(cols, attrNames...)

	var b strings.Builder
	b.WriteString(strings.Join(cols, "|"))
	for _, p := range players {
		row := []string{
			strconv.FormatInt(p.UID, 10),
			chatTableFieldSafe(p.Name),
			chatTableFieldSafe(p.Club),
			p.Age,
			strings.Join(p.ShortPositions, ","),
			strconv.Itoa(p.Overall),
			strconv.Itoa(p.CA),
			chatTableFieldSafe(p.BestRoleOverall),
			strconv.FormatInt(p.TransferValueAmount, 10),
			strconv.FormatInt(p.WageAmount, 10),
			chatTableFieldSafe(p.Nationality),
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

// --- search_players tool (ticket 02) ---

type chatSearchPlayersArgs struct {
	Position    string   `json:"position"`
	Club        string   `json:"club"`
	Division    string   `json:"division"`
	Nationality string   `json:"nationality"`
	MinAge      int      `json:"minAge"`
	MaxAge      int      `json:"maxAge"`
	MinOverall  int      `json:"minOverall"`
	MaxOverall  int      `json:"maxOverall"`
	Attributes  []string `json:"attributes"`
	SortBy      string   `json:"sortBy"`
	Limit       int      `json:"limit"`
}

var chatSearchPlayersToolSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"position":    map[string]any{"type": "string", "description": "Short position code to filter by, e.g. \"ST\", \"CB\", \"DL\". Omit for any position."},
		"club":        map[string]any{"type": "string", "description": "Exact club name to scope the search to that club's own players. Omit to search the wider transfer market."},
		"division":    map[string]any{"type": "string", "description": "League/division name to filter by. Omit for any division."},
		"nationality": map[string]any{"type": "string", "description": "Player nationality to filter by. Omit for any nationality."},
		"minAge":      map[string]any{"type": "number", "description": "Minimum age. 0 or omitted means no minimum."},
		"maxAge":      map[string]any{"type": "number", "description": "Maximum age. 0 or omitted means no maximum."},
		"minOverall":  map[string]any{"type": "number", "description": "Minimum overall rating. 0 or omitted means no minimum."},
		"maxOverall":  map[string]any{"type": "number", "description": "Maximum overall rating. 0 or omitted means no maximum."},
		"attributes": map[string]any{
			"type":        "array",
			"items":       map[string]any{"type": "string"},
			"description": "Up to 5 attribute names (long form, e.g. \"Finishing\", \"Pace\") to include in each result's attribute values.",
		},
		"sortBy": map[string]any{
			"type":        "string",
			"enum":        []string{"overall_desc", "overall_asc", "age_asc", "age_desc"},
			"description": "Sort order for results. Defaults to overall_desc.",
		},
		"limit": map[string]any{"type": "number", "description": "Max results to return. Defaults to 15; capped at 100 regardless of what's requested."},
	},
}

func chatSearchPlayers(datasetID string, args chatSearchPlayersArgs) []ChatPlayerSummary {
	players, _, found := GetPlayerData(datasetID)
	if !found {
		return nil
	}
	players = RecalculateAllPlayersRatings(players)

	matches := make([]Player, 0, 32)
	for i := range players {
		player := players[i]
		if args.Position != "" {
			hasPosition := false
			for _, pos := range player.ShortPositions {
				if strings.EqualFold(pos, args.Position) {
					hasPosition = true
					break
				}
			}
			if !hasPosition {
				continue
			}
		}
		if args.Club != "" && !strings.EqualFold(player.Club, args.Club) {
			continue
		}
		if args.Division != "" && !strings.EqualFold(player.Division, args.Division) {
			continue
		}
		if args.Nationality != "" && !strings.EqualFold(player.Nationality, args.Nationality) {
			continue
		}
		if args.MinAge > 0 || args.MaxAge > 0 {
			playerAge, ageErr := strconv.Atoi(player.Age)
			if ageErr != nil {
				continue
			}
			if args.MinAge > 0 && playerAge < args.MinAge {
				continue
			}
			if args.MaxAge > 0 && playerAge > args.MaxAge {
				continue
			}
		}
		if args.MinOverall > 0 && player.Overall < args.MinOverall {
			continue
		}
		if args.MaxOverall > 0 && player.Overall > args.MaxOverall {
			continue
		}
		matches = append(matches, player)
	}

	switch args.SortBy {
	case "overall_asc":
		sort.Slice(matches, func(i, j int) bool { return matches[i].Overall < matches[j].Overall })
	case "age_asc":
		sort.Slice(matches, func(i, j int) bool { return matches[i].Age < matches[j].Age })
	case "age_desc":
		sort.Slice(matches, func(i, j int) bool { return matches[i].Age > matches[j].Age })
	default:
		sort.Slice(matches, func(i, j int) bool { return matches[i].Overall > matches[j].Overall })
	}

	limit := args.Limit
	if limit <= 0 {
		limit = chatbotSearchDefaultLimit
	}
	if limit > chatbotSearchMaxLimit {
		limit = chatbotSearchMaxLimit
	}
	if len(matches) > limit {
		matches = matches[:limit]
	}

	summaries := make([]ChatPlayerSummary, len(matches))
	for i, player := range matches {
		summaries[i] = buildChatPlayerSummary(player, args.Attributes)
	}
	return summaries
}

// --- compare_squads tool (ticket 02) ---

type chatCompareSquadsArgs struct {
	ClubA string `json:"clubA"`
	ClubB string `json:"clubB"`
}

var chatCompareSquadsToolSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"clubA": map[string]any{"type": "string", "description": "First club's exact name."},
		"clubB": map[string]any{"type": "string", "description": "Second club's exact name."},
	},
	"required": []string{"clubA", "clubB"},
}

// chatPositionBuckets groups short position codes into six broad buckets for a rough
// squad-composition comparison — not a precise tactical breakdown, just enough for
// "how do these two squads compare" at a glance. A player's primary (first-listed)
// short position decides its bucket, so a player is never double-counted.
var chatPositionBucketOrder = []string{"GK", "CB", "FB", "CM", "W", "ST"}

func chatPositionBucket(shortPosition string) string {
	switch strings.ToUpper(shortPosition) {
	case "GK":
		return "GK"
	case "DC":
		return "CB"
	case "DL", "DR", "WBL", "WBR":
		return "FB"
	case "DM", "MC", "AMC":
		return "CM"
	case "AML", "AMR", "ML", "MR":
		return "W"
	case "ST":
		return "ST"
	default:
		return ""
	}
}

type chatSquadBucketStat struct {
	AvgOverall  int
	Count       int
	BestOverall int
}

// computeSquadBuckets buckets a club's players and computes per-bucket average/best
// overall — shared between the compare_squads tool and render_chart's team_bar
// template (both new to this file; not a reuse of any pre-existing dialog's logic).
func computeSquadBuckets(datasetID, club string) (map[string]chatSquadBucketStat, int, bool) {
	players, _, found := GetPlayerData(datasetID)
	if !found {
		return nil, 0, false
	}
	sums := make(map[string]int)
	counts := make(map[string]int)
	bests := make(map[string]int)
	for i := range players {
		player := players[i]
		if !strings.EqualFold(player.Club, club) {
			continue
		}
		if len(player.ShortPositions) == 0 {
			continue
		}
		bucket := chatPositionBucket(player.ShortPositions[0])
		if bucket == "" {
			continue
		}
		sums[bucket] += player.Overall
		counts[bucket]++
		if player.Overall > bests[bucket] {
			bests[bucket] = player.Overall
		}
	}
	if len(counts) == 0 {
		return nil, 0, false
	}
	stats := make(map[string]chatSquadBucketStat, len(chatPositionBucketOrder))
	bestXISum, bestXICount := 0, 0
	for _, bucket := range chatPositionBucketOrder {
		if counts[bucket] == 0 {
			continue
		}
		stats[bucket] = chatSquadBucketStat{
			AvgOverall:  sums[bucket] / counts[bucket],
			Count:       counts[bucket],
			BestOverall: bests[bucket],
		}
		bestXISum += bests[bucket]
		bestXICount++
	}
	bestXI := 0
	if bestXICount > 0 {
		bestXI = bestXISum / bestXICount
	}
	return stats, bestXI, true
}

func renderCompareSquadsTable(clubA string, statsA map[string]chatSquadBucketStat, bestXIA int, clubB string, statsB map[string]chatSquadBucketStat, bestXIB int) string {
	var b strings.Builder
	b.WriteString("position|" + chatTableFieldSafe(clubA) + "_avgOvr|" + chatTableFieldSafe(clubA) + "_depth|" + chatTableFieldSafe(clubB) + "_avgOvr|" + chatTableFieldSafe(clubB) + "_depth")
	for _, bucket := range chatPositionBucketOrder {
		a := statsA[bucket]
		bb := statsB[bucket]
		b.WriteByte('\n')
		b.WriteString(fmt.Sprintf("%s|%d|%d|%d|%d", bucket, a.AvgOverall, a.Count, bb.AvgOverall, bb.Count))
	}
	b.WriteByte('\n')
	b.WriteString(fmt.Sprintf("bestXI (avg of each bucket's best player)|%d||%d|", bestXIA, bestXIB))
	return b.String()
}

// --- get_managed_squad tool (ticket 02, 03, 09) ---

var chatGetManagedSquadToolSchema = map[string]any{
	"type":       "object",
	"properties": map[string]any{},
}

// getManagedSquadPlayers is a small, self-contained squad lookup — deliberately not a
// call into who_to_sign.go's getSquadForTeam, keeping the "fully independent" boundary
// between the two features unambiguous even though the lookup itself is trivial.
func getManagedSquadPlayers(datasetID, club string) ([]Player, bool) {
	players, _, found := GetPlayerData(datasetID)
	if !found {
		return nil, false
	}
	squad := make([]Player, 0)
	for i := range players {
		if players[i].Club == club {
			squad = append(squad, players[i])
		}
	}
	return squad, true
}

// resolveBasedInNation reads Player.BasedIn off the squad's own players (ticket 03) —
// a real, pre-existing FM data field, not a heuristic. Majority-vote across the squad
// is a defensive fallback for the rare case of data inconsistency; every player at a
// given club should carry the same value in practice.
func resolveBasedInNation(squad []Player) string {
	counts := make(map[string]int)
	for i := range squad {
		if squad[i].BasedIn != "" {
			counts[squad[i].BasedIn]++
		}
	}
	best, bestCount := "", 0
	for nation, count := range counts {
		if count > bestCount {
			best, bestCount = nation, count
		}
	}
	return best
}

// chatGetManagedSquad returns {basedInNation, players[]} fresh on every call — no
// intra-turn memoization (ticket 09): it's a cheap in-memory lookup, not worth caching.
func chatGetManagedSquad(datasetID, club string) (basedInNation string, summaries []ChatPlayerSummary, ok bool) {
	squad, found := getManagedSquadPlayers(datasetID, club)
	if !found {
		return "", nil, false
	}
	basedInNation = resolveBasedInNation(squad)
	summaries = make([]ChatPlayerSummary, len(squad))
	for i, player := range squad {
		summaries[i] = buildChatPlayerSummary(player, nil)
	}
	return basedInNation, summaries, true
}

// --- render_chart tool (ticket 02, 04) ---

type chatRenderChartArgs struct {
	Template   string   `json:"template"`
	PlayerUIDs []int64  `json:"playerUids"`
	Clubs      []string `json:"clubs"`
}

var chatRenderChartToolSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"template": map[string]any{
			"type":        "string",
			"enum":        []string{"player_radar", "team_bar"},
			"description": "Which chart template to render.",
		},
		"playerUids": map[string]any{
			"type":        "array",
			"items":       map[string]any{"type": "integer"},
			"description": "For player_radar only: 1-2 player uids (from a tool result this conversation) to compare.",
		},
		"clubs": map[string]any{
			"type":        "array",
			"items":       map[string]any{"type": "string"},
			"description": "For team_bar only: exactly 2 club names to compare.",
		},
	},
	"required": []string{"template"},
}

// chatRadarAttributes are six representative attributes (across physical/technical/
// mental categories) shown on the player_radar template — not the full attribute
// breakdown PlayerComparisonDialog.vue shows, just enough for an at-a-glance shape.
var chatRadarLabels = []string{"Pace", "Passing", "Tackling", "Finishing", "Positioning", "Stamina"}
var chatRadarCodes = []string{"Pac", "Pas", "Tck", "Fin", "Pos", "Sta"}
var chatRadarBorderColors = []string{"#3b82f6", "#ef4444"}
var chatRadarBgColors = []string{"rgba(59,130,246,0.15)", "rgba(239,68,68,0.15)"}

// buildPlayerRadarChart re-derives the plotted values fresh from the live dataset by
// uid (ticket 04) — never trusts model-supplied numbers.
func buildPlayerRadarChart(datasetID string, uids []int64) *ChatChart {
	players, _, found := GetPlayerData(datasetID)
	if !found {
		return nil
	}
	byUID := make(map[int64]Player, len(players))
	for i := range players {
		byUID[players[i].UID] = players[i]
	}

	var datasets []map[string]any
	for i, uid := range uids {
		if i >= 2 {
			break
		}
		player, ok := byUID[uid]
		if !ok {
			continue
		}
		data := make([]int, len(chatRadarCodes))
		for j, code := range chatRadarCodes {
			data[j] = player.NumericAttributes[code]
		}
		datasets = append(datasets, map[string]any{
			"label":           player.Name,
			"data":            data,
			"borderColor":     chatRadarBorderColors[i],
			"backgroundColor": chatRadarBgColors[i],
		})
	}
	if len(datasets) == 0 {
		return nil
	}
	return &ChatChart{
		Template: "player_radar",
		Data:     map[string]any{"labels": chatRadarLabels, "datasets": datasets},
	}
}

var chatBarColors = []string{"#3b82f6", "#ef4444"}

// buildTeamBarChart re-derives per-position averages fresh via computeSquadBuckets —
// same grounding-over-trusting-the-model rationale as buildPlayerRadarChart.
func buildTeamBarChart(datasetID string, clubs []string) *ChatChart {
	if len(clubs) != 2 {
		return nil
	}
	statsA, _, okA := computeSquadBuckets(datasetID, clubs[0])
	statsB, _, okB := computeSquadBuckets(datasetID, clubs[1])
	if !okA || !okB {
		return nil
	}
	dataA := make([]int, len(chatPositionBucketOrder))
	dataB := make([]int, len(chatPositionBucketOrder))
	for i, bucket := range chatPositionBucketOrder {
		dataA[i] = statsA[bucket].AvgOverall
		dataB[i] = statsB[bucket].AvgOverall
	}
	datasets := []map[string]any{
		{"label": clubs[0], "data": dataA, "backgroundColor": chatBarColors[0]},
		{"label": clubs[1], "data": dataB, "backgroundColor": chatBarColors[1]},
	}
	return &ChatChart{
		Template: "team_bar",
		Data:     map[string]any{"labels": chatPositionBucketOrder, "datasets": datasets},
	}
}

// --- System prompt (ticket 08) ---

func buildChatbotSystemPrompt(managedClub string) string {
	var b strings.Builder
	b.WriteString("You are an Association Football Manager's data-savvy scouting and tactical assistant, embedded as a chat widget in the club's data dashboard. The manager's club is " + managedClub + ".\n\n")
	b.WriteString("Tools available to you:\n")
	b.WriteString("- search_players: search the transfer market, or a specific club's players (set club to scope it; omit club for the wider market). You may call it multiple times per turn, including chaining calls where a value from one search feeds the next — e.g. to find a player better than a specific club's best at a position, first search that club at that position to find their overall, then search again with club omitted and minOverall set to that value.\n")
	b.WriteString("- compare_squads: get per-position average overall, depth, and a best-XI figure for two clubs — use this for squad composition comparisons instead of eyeballing full rosters yourself.\n")
	b.WriteString("- get_managed_squad: get the manager's own current squad plus the nation their club is based in. Always call this before answering a homegrown-talent or tactic-fit question.\n")
	b.WriteString("- render_chart: optionally render a chart (player_radar for 1-2 players' attributes, team_bar for a two-club squad comparison) to accompany your answer. Only call it when a visual genuinely helps — most answers don't need one.\n\n")
	b.WriteString("Instructions:\n")
	b.WriteString("- Never state a player's stats or attributes that didn't come from a tool result this conversation — do not invent numbers.\n")
	b.WriteString("- Use Football Manager terminology and this app's position/attribute naming conventions, not generic football commentary or FIFA-style ratings — PAC, SHO, PAS, DRI, DEF, and PHY do not exist here and must never appear in your reasoning.\n")
	b.WriteString("- Write like a natural, concise chat reply, not a formal report — a sentence or a short paragraph is usually enough, backed by specifics.\n")
	b.WriteString("- For \"what tactic would fit this squad\" style questions, ground every formation or role claim in a specific attribute or position strength from get_managed_squad's data — never give generic tactical advice untethered from the actual squad.\n")
	b.WriteString("- To mention a specific player by name in your reply, write the token [[player:UID]] in place of their name (e.g. \"[[player:104829]] would be a strong pick\") and include them in referencedPlayers with their uid and name — the interface turns this into a clickable link. Only cite players who came from a tool result this conversation.\n")
	b.WriteString("- If a question is clearly unrelated to scouting or squad analysis, briefly acknowledge that and steer the conversation back to what you can help with.\n")
	b.WriteString("- Respond only in the structured format provided.\n")
	return b.String()
}

// --- Handler (ticket 06, 10) ---

func chatbotHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	datasetID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/chatbot/"), "/")
	if datasetID == "" {
		http.Error(w, "Dataset ID is missing in the request path", http.StatusBadRequest)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Error parsing request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if len(req.History) == 0 || req.History[len(req.History)-1].Role != "user" {
		writeChatbotPreStreamError(w, r, http.StatusBadRequest, "history must be non-empty and end with a user message")
		return
	}
	userTurns := 0
	for _, turn := range req.History {
		if turn.Role == "user" {
			userTurns++
		}
	}
	if userTurns > chatbotMaxHistoryTurns {
		writeChatbotPreStreamError(w, r, http.StatusBadRequest, "this conversation has reached its turn limit — start a New Chat to continue")
		return
	}

	apiKey := r.Header.Get("X-OpenAI-Api-Key")
	if apiKey == "" {
		apiKey = whoToSignEnvAPIKey()
	}
	if apiKey == "" {
		writeChatbotPreStreamError(w, r, http.StatusBadRequest, "no API key provided — add one in Settings")
		return
	}

	team, found := getManagedTeam(datasetID)
	if !found || team.Club == "" {
		writeChatbotPreStreamError(w, r, http.StatusBadRequest, "no managed team set for this dataset")
		return
	}
	if _, _, found := GetPlayerData(datasetID); !found {
		writeChatbotPreStreamError(w, r, http.StatusNotFound, "dataset not found")
		return
	}

	logInfo(ctx, "Processing chatbot request", "dataset_id", datasetID, "club", team.Club, "turns", userTurns)

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

	client := openai.NewClient(option.WithAPIKey(apiKey))
	systemPrompt := buildChatbotSystemPrompt(team.Club)
	result, chart, chatErr := runChatbotLoop(ctx, &client, datasetID, team.Club, systemPrompt, req.History, sendStatus)
	if chatErr != nil {
		logWarn(ctx, "chatbot request failed", "dataset_id", datasetID, "status", chatErr.status, "error", chatErr.message)
		writeSSEEvent(w, "error", map[string]string{"code": strconv.Itoa(chatErr.status), "message": chatErr.message})
		if canFlush {
			flusher.Flush()
		}
		return
	}

	writeSSEEvent(w, "done", ChatDoneEvent{
		Text:              result.Text,
		ReferencedPlayers: result.ReferencedPlayers,
		Chart:             chart,
	})
	if canFlush {
		flusher.Flush()
	}
}

func writeSSEEvent(w http.ResponseWriter, event string, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
}

func writeChatbotPreStreamError(w http.ResponseWriter, r *http.Request, status int, message string) {
	setCORSHeaders(w, r)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// getManagedTeam reads the dataset's managed team from configStorage — reuses
// managed_team.go's ManagedTeam type and storage key, which is pure data-storage
// plumbing shared with Scout Report/Who to Sign, not scouting business logic.
func getManagedTeam(datasetID string) (ManagedTeam, bool) {
	if configStorage == nil {
		return ManagedTeam{}, false
	}
	data, err := configStorage.RetrieveConfig(managedTeamStorageKey(datasetID))
	if err != nil {
		return ManagedTeam{}, false
	}
	var team ManagedTeam
	if err := json.Unmarshal(data, &team); err != nil {
		return ManagedTeam{}, false
	}
	return team, true
}

var chatbotToolFriendlyLabels = map[string]string{
	"search_players":    "Searching players…",
	"compare_squads":    "Comparing squads…",
	"get_managed_squad": "Reviewing your squad…",
	"render_chart":      "Building chart…",
}

// runChatbotLoop drives one turn's Responses API tool-calling loop: declares all four
// tools, executes each in-process on function_call, and returns once the model produces
// a final structured answer — capped at chatbotMaxToolRoundsPerTurn rounds (ticket 01).
// Conversation history is rebuilt from req.History on every call (ticket 01: stateless,
// no PreviousResponseID carried across HTTP turns — only within this one turn's loop).
func runChatbotLoop(ctx context.Context, client *openai.Client, datasetID, managedClub, systemPrompt string, history []ChatHistoryTurn, sendStatus func(string)) (*chatModelResponse, *ChatChart, *whoToSignError) {
	items := make([]responses.ResponseInputItemUnionParam, 0, len(history))
	for _, turn := range history {
		role := responses.EasyInputMessageRoleUser
		if turn.Role == "assistant" {
			role = responses.EasyInputMessageRoleAssistant
		}
		items = append(items, responses.ResponseInputItemParamOfMessage(turn.Text, role))
	}

	tools := []responses.ToolUnionParam{
		{OfFunction: &responses.FunctionToolParam{
			Name:        "search_players",
			Description: openai.String("Search the transfer market or a specific club's players. Returns a pipe-delimited table (header row first): uid|name|club|age|pos|ovr|ca|role|value|wage|nat, plus one column per requested attribute."),
			Parameters:  chatSearchPlayersToolSchema,
		}},
		{OfFunction: &responses.FunctionToolParam{
			Name:        "compare_squads",
			Description: openai.String("Compare two clubs' squad composition. Returns a pipe-delimited table of per-position average overall and depth for both clubs, plus a best-XI figure for each."),
			Parameters:  chatCompareSquadsToolSchema,
		}},
		{OfFunction: &responses.FunctionToolParam{
			Name:        "get_managed_squad",
			Description: openai.String("Get the manager's own current squad (as a pipe-delimited table, same column format as search_players) plus the nation their club is based in."),
			Parameters:  chatGetManagedSquadToolSchema,
		}},
		{OfFunction: &responses.FunctionToolParam{
			Name:        "render_chart",
			Description: openai.String("Render a chart to accompany your answer. player_radar needs playerUids (1-2); team_bar needs clubs (exactly 2)."),
			Parameters:  chatRenderChartToolSchema,
		}},
	}
	textFormat := responses.ResponseTextConfigParam{
		Format: func() responses.ResponseFormatTextConfigUnionParam {
			schema := responses.ResponseFormatTextJSONSchemaConfigParam{
				Name:   "chatbot_response",
				Schema: chatResponseSchema,
				Strict: openai.Bool(true),
			}
			return responses.ResponseFormatTextConfigUnionParam{OfJSONSchema: &schema}
		}(),
	}

	params := responses.ResponseNewParams{
		Model:        chatbotModel,
		Instructions: openai.String(systemPrompt),
		Input:        responses.ResponseNewParamsInputUnion{OfInputItemList: items},
		Tools:        tools,
		Text:         textFormat,
	}

	resp, apiErr := callResponsesWithRetry(ctx, client, params)
	if apiErr != nil {
		return nil, nil, apiErr
	}

	var lastChart *ChatChart

	for round := 0; ; round++ {
		var toolCalls []responses.ResponseFunctionToolCall
		for i := range resp.Output {
			if resp.Output[i].Type == "function_call" {
				toolCalls = append(toolCalls, resp.Output[i].AsFunctionCall())
			}
		}
		if len(toolCalls) == 0 {
			var result chatModelResponse
			outputText := resp.OutputText()
			if err := json.Unmarshal([]byte(outputText), &result); err != nil {
				logWarn(ctx, "chatbot final response could not be parsed",
					"error", err.Error(),
					"status", resp.Status,
					"output_text_len", len(outputText),
				)
				return nil, nil, &whoToSignError{http.StatusInternalServerError, "the assistant's response could not be parsed"}
			}
			return &result, lastChart, nil
		}
		if round >= chatbotMaxToolRoundsPerTurn {
			break
		}

		seenLabels := make(map[string]bool)
		for _, toolCall := range toolCalls {
			if !seenLabels[toolCall.Name] {
				seenLabels[toolCall.Name] = true
				if label, ok := chatbotToolFriendlyLabels[toolCall.Name]; ok {
					sendStatus(label)
				}
			}
		}

		outputItems := make([]responses.ResponseInputItemUnionParam, len(toolCalls))
		for i, toolCall := range toolCalls {
			outputText := dispatchChatbotTool(datasetID, managedClub, toolCall.Name, toolCall.Arguments, &lastChart)
			outputItems[i] = responses.ResponseInputItemUnionParam{
				OfFunctionCallOutput: &responses.ResponseInputItemFunctionCallOutputParam{
					CallID: toolCall.CallID,
					Output: responses.ResponseInputItemFunctionCallOutputOutputUnionParam{
						OfString: openai.String(outputText),
					},
				},
			}
		}

		nextParams := responses.ResponseNewParams{
			Model:              chatbotModel,
			PreviousResponseID: openai.String(resp.ID),
			Tools:              tools,
			Text:               textFormat,
			Input:              responses.ResponseNewParamsInputUnion{OfInputItemList: outputItems},
		}
		resp, apiErr = callResponsesWithRetry(ctx, client, nextParams)
		if apiErr != nil {
			return nil, nil, apiErr
		}
	}

	return nil, nil, &whoToSignError{
		http.StatusGatewayTimeout,
		"the assistant couldn't settle on an answer within the allowed search attempts — try a narrower question and try again.",
	}
}

// dispatchChatbotTool executes one tool call and returns its result as a string ready
// to hand back to the model. lastChart is updated in place when render_chart succeeds.
func dispatchChatbotTool(datasetID, managedClub, name, argsJSON string, lastChart **ChatChart) string {
	switch name {
	case "search_players":
		var args chatSearchPlayersArgs
		_ = json.Unmarshal([]byte(argsJSON), &args)
		return renderChatPlayerTable(chatSearchPlayers(datasetID, args))

	case "compare_squads":
		var args chatCompareSquadsArgs
		_ = json.Unmarshal([]byte(argsJSON), &args)
		statsA, bestXIA, okA := computeSquadBuckets(datasetID, args.ClubA)
		statsB, bestXIB, okB := computeSquadBuckets(datasetID, args.ClubB)
		if !okA || !okB {
			return "One or both clubs have no players in this dataset — check the club names."
		}
		return renderCompareSquadsTable(args.ClubA, statsA, bestXIA, args.ClubB, statsB, bestXIB)

	case "get_managed_squad":
		basedInNation, summaries, ok := chatGetManagedSquad(datasetID, managedClub)
		if !ok {
			return "Could not load the managed squad."
		}
		return "basedInNation: " + basedInNation + "\n" + renderChatPlayerTable(summaries)

	case "render_chart":
		var args chatRenderChartArgs
		_ = json.Unmarshal([]byte(argsJSON), &args)
		var chart *ChatChart
		switch args.Template {
		case "player_radar":
			chart = buildPlayerRadarChart(datasetID, args.PlayerUIDs)
		case "team_bar":
			chart = buildTeamBarChart(datasetID, args.Clubs)
		}
		if chart == nil {
			return "Chart could not be rendered — check the player uids or club names."
		}
		*lastChart = chart
		return "Chart rendered."

	default:
		return "Unknown tool."
	}
}
