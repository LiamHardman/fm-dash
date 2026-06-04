# Feature Suggestions

Suggested features for FM-Dash based on a high-level review of the codebase. Where a suggestion overlaps with an existing roadmap item in `roadmap.md`, this is noted with a reference.

---

## 1. Player Comparison Tool

**Summary:** Side-by-side comparison of 2–4 players with a spider/radar chart overlay and a tabular attribute diff.

**Detail:**
Users regularly need to choose between two or more transfer targets. Today they have to open each player's detail dialog in sequence and mentally compare them. A dedicated comparison view would let users pin players (from the main table, wishlist, or search) and see all their attributes, role suitabilities, percentiles, and financials in a single view.

The radar chart would overlay each player's attribute profile using a common axis scale, making strengths and weaknesses immediately visible. Below the chart, a diff table would highlight which attributes each player leads in (e.g. Player A has higher Pace, Player B has higher Positioning). A "winner" summary row could call out the better player for each selected role.

**Implementation notes:**
- All data is already present on the `Player` protobuf model — no new API work needed.
- The existing `PlayerDetailDialog` holds most of the display logic; this would be a new route/modal that accepts an array of player UIDs.
- A charting library (e.g. Chart.js or ECharts, both common in Vue ecosystems) would handle the radar overlay.
- Entry points: "Add to comparison" button on player rows and the player detail dialog; a persistent comparison tray (like a cookie bar) shows pinned players.

**Effort estimate:** Medium — mostly frontend work, no backend changes.

---

## 2. Squad Depth Analyzer

**Summary:** For a selected team, visualise positional depth as a layered formation grid and surface coverage gaps.

**Detail:**
Currently the Nations page shows formation-level squad composition, but there's no equivalent tool for club teams. A Squad Depth Analyzer would show, for any team in the dataset, how many players are available at each position and how their quality stacks up (starter / rotation / depth).

The view would render a pitch split into positional zones. Each zone would show the top-rated player for that slot, plus a count badge for backups. Zones where the team has fewer than two viable options would be flagged in amber/red. A sidebar would list recommended signings from the same dataset to fill each gap, linking directly to the Upgrade Finder filtered to that position.

This would be especially useful for players who want to assess a club before joining or identify the weakest areas in their own squad mid-save.

**Implementation notes:**
- Team roster data is already aggregated; position grouping logic exists in the player model (`position_groups`, `short_positions`).
- The pitch display component can be extended — it already supports player positioning by role.
- "Gap detection" is a threshold check: fewer than N players in a position group with an overall above a configurable minimum.
- Feeds naturally into the Upgrade Finder (roadmap #67 is adjacent — defining a team from a second dataset).

**Effort estimate:** Medium — extends existing pitch display, modest backend filtering work.

---

## 3. Custom Rating Builder

**Summary:** Let users define their own overall formula by weighting individual attributes, then apply it as a sortable column in the player table.

**Detail:**
FM-Dash already offers MBR, Total Stats, and role-specific overalls. Different playstyles demand different priorities — a pressing manager cares about Stamina and Work Rate; a tiki-taka manager cares about Short Passing and First Touch. No pre-built formula covers every use case.

The Custom Rating Builder would be a modal (accessible from Settings or the table column picker) that lists all numeric attributes and lets the user assign a weight (0–5 stars, or a numeric slider) to each. The resulting formula is normalised to a 0–100 scale and applied as a new column in the player table — sortable and filterable like any other rating. Users could save and name multiple presets (e.g. "My Gegenpressing Rating", "Bargain Keeper Score").

**Implementation notes:**
- The MBR calculation pattern (Go backend, web worker) is the right model to follow.
- Presets can be stored in `localStorage` on the frontend — no auth or backend persistence needed initially.
- The backend would need a new endpoint that accepts a weight map and player UIDs, or the calculation could move to a web worker entirely (all numeric attributes are already returned to the frontend).
- Overlaps loosely with roadmap #51 (new total stat score) and #63 (docs on rating calculations) — this would supersede the need for multiple static scores.

**Effort estimate:** Medium — the computation pattern is established; the main work is the weights UI and preset management.

---

## 4. Age / Potential Scatter Plot

**Summary:** A 2D chart plotting all filtered players by age (x-axis) vs. overall or potential (y-axis), with interactive hover and click-through to player details.

**Detail:**
The player table is excellent for precise filtering but makes it hard to grasp the shape of a dataset at a glance. A scatter plot would make the "diamond in the rough" pattern immediately visible: a young player with high potential sitting in the bottom-left quadrant is a classic Wonderkid buy; an older player with high current ability in the top-right is a proven veteran.

Each dot would be coloured by position group (GK, DEF, MID, ATT) and sized by transfer value. Hovering shows a tooltip with name, club, and key stats. Clicking opens the player detail dialog. The chart would respect the active filter state so users can narrow it down (e.g. "only Central Midfielders in the Championship") before interpreting the chart.

A toggle would switch between "Overall vs. Age" and "Potential vs. Age" (where potential data is available), and a second toggle could overlay the user's wishlist players as highlighted dots.

**Implementation notes:**
- No new backend work needed — all data is already returned.
- ECharts or Chart.js scatter supports this pattern well and is already likely available via Quasar's ecosystem.
- Can be added as a view toggle on the existing dataset page (table / chart) rather than a separate route.
- Overlaps with Wonderkids (which already segments by age and potential) — this generalises that concept visually.

**Effort estimate:** Low-Medium — purely a frontend visualisation over existing data.

---

## 5. Multi-Dataset Comparison (Save File Diff)

**Summary:** Upload two datasets from different points in a save and see what changed — player development, new arrivals, departures, and value shifts.

**Detail:**
FM players often upload a snapshot at the start of a season and again mid-season or at the end. Right now each dataset is viewed in isolation. A diff view would let users select two datasets and see:

- **Improved players:** attributes that grew since the first snapshot, with delta values.
- **Declined players:** regression due to age, injuries, or form.
- **New arrivals / departures:** players who appear in one dataset but not the other, grouped by club or league.
- **Value shifts:** players whose transfer value increased or decreased the most (useful for identifying assets to sell).

The diff would be scoped — users can focus on a specific team, league, or position group rather than the whole dataset.

**Implementation notes:**
- Requires two dataset IDs to be selected; the diff logic can live on the backend (Go diff endpoint) or be computed on the frontend by joining on player UID.
- UID stability across FM exports is a prerequisite — worth validating this assumption first.
- Authentication (roadmap #54) would make this much smoother since users could link datasets together, but it could also work by simply selecting two datasets from the existing list.
- Aligns with roadmap #54's goal of linking datasets together.

**Effort estimate:** Medium-High — UID stability needs validating; diff logic and UI are new surface area.

---

## 6. Squad Builder / Tactic Planner

**Summary:** An interactive formation canvas where users drag players into positions, score the lineup by role suitability, and receive improvement suggestions.

**Detail:**
The existing pitch display visualises a formation statically. This feature would make it interactive: users pick a formation (4-3-3, 4-2-3-1, etc.), drag players from their dataset into each position slot, and see a live score for how well each player fits their assigned role (using the existing `roleSpecificOveralls` data).

The overall lineup score would be an aggregate of individual role fits. Underperforming slots would be highlighted with a "better options" popover showing the top 3 alternatives from the dataset. Users could save lineups as named tactics and export them.

A "Auto-fill" button would populate the best-fit eleven for a selected formation from the dataset — useful as a starting point.

**Implementation notes:**
- Overlaps directly with roadmap #60 ("Create a team" page) — this is a detailed spec for that item.
- `roleSpecificOveralls` is already computed per player — the scoring logic is available.
- Formation templates can be a static config (JSON) — FM has a defined set of formations.
- Drag-and-drop: Quasar has a drag-and-drop directive; alternatively Vue Draggable is a common choice.
- Auto-fill is an assignment problem (bipartite matching) — a greedy algorithm is sufficient for this use case.

**Effort estimate:** High — significant frontend complexity; backend work is mostly already done.

---

## 7. Player Archetype Clustering

**Summary:** Automatically group players in a dataset into archetypes (e.g. "Deep-Lying Playmaker", "Press-Forward Striker") based on their attribute profiles, and let users search for players by archetype.

**Detail:**
Finding "a player like X but cheaper" is one of the most common FM scouting tasks. Archetype clustering solves this by grouping players with similar attribute fingerprints. Instead of defining complex filter criteria, a user could browse archetypes, find the group their current player belongs to, and see all statistically similar players ranked by cost-effectiveness.

Archetypes would be labelled with a descriptive name derived from the dominant attributes and best role. Each archetype page would show the typical attribute profile and list member players, sortable by age, value, or rating. A "find similar" button on any player detail dialog would jump to that player's archetype filtered to the current dataset.

**Implementation notes:**
- Overlaps with roadmap #58 ("Playstyle display for players") — this extends that concept into a browsable taxonomy.
- Clustering could be done at dataset upload time on the backend (k-means or similar over numeric attributes), or approximated heuristically using the existing role suitability scores.
- For the heuristic approach: group players by their top 2 role suitabilities — already computed.
- No ML infrastructure required for the heuristic path; full clustering would need a Go ML library or an offline pre-computation step.

**Effort estimate:** Medium (heuristic) / High (ML clustering).

---

## 8. Exportable PDF Scouting Report

**Summary:** Generate a formatted, shareable one-page scouting report PDF for any player, containing key stats, role suitabilities, percentile rankings, and a written summary.

**Detail:**
CSV export is already supported for bulk data, but there's no way to share a polished report on a single player. A PDF scouting report would be useful for club managers sharing target intel with their backroom staff, or community members posting player reviews.

The report would include: player photo (regen face if available), key attributes in a visual layout, top 3 role suitabilities with scores, percentile rankings for key stats relative to the division, a "strengths / weaknesses" callout derived from attributes above/below a threshold, and financial summary (value, wage).

An optional AI-generated written summary could be toggled on for users who want a narrative paragraph — this links naturally to roadmap #59.

**Implementation notes:**
- `jsPDF` or `html2canvas` + `jsPDF` is the standard Vue/browser approach — no backend needed.
- The player detail dialog already contains all the required display elements; the PDF is largely a print-styled version of that view.
- AI summary (roadmap #59) could be wired in as an optional enrichment if/when that feature lands.

**Effort estimate:** Low-Medium — mostly a styled print template; AI integration is optional and incremental.

---

## 9. Pinned Filter Presets

**Summary:** Let users save and name the current filter state as a reusable preset, with one-click restore.

**Detail:**
The filter panel supports a rich combination of criteria (position, league, age range, nationality, division, attribute thresholds, etc.). Building a specific filter from scratch every session is friction. Presets would let users save a named snapshot of any filter state (e.g. "Cheap U21 Strikers", "Liga MX GKs", "Free Agent Defenders") and reload it instantly.

Presets would appear as chips below the filter panel header. Clicking a chip applies all its criteria instantly. Users can delete, rename, or overwrite presets. Presets are stored in `localStorage` so they persist across sessions without requiring authentication.

**Implementation notes:**
- The filter state is already a reactive Pinia store — serialising it to JSON for storage is straightforward.
- No backend changes needed; `localStorage` is sufficient.
- Small UI addition: a "Save preset" button in the filter panel header, and a chip list above or below the filter controls.

**Effort estimate:** Low — well-bounded frontend-only feature.

---

## 10. "Similar Players" on Player Detail

**Summary:** At the bottom of the player detail dialog, show the top 5 most attribute-similar players in the current dataset.

**Detail:**
When evaluating a player, it's immediately useful to know who else in the dataset has a comparable profile — whether to find a backup option or benchmark the player's value. "Similar Players" would compute a similarity score (Euclidean distance over normalised numeric attributes) against all other players in the dataset and surface the top 5 matches.

Each similar player card would show name, club, age, overall, and transfer value, with a one-click link to their own detail dialog. An optional "similarity type" toggle could switch between "similar overall profile" and "similar in this role" (using role-specific attribute subsets).

**Implementation notes:**
- Similarity computation can happen in a web worker at dialog-open time — the attribute vectors are already in the frontend store.
- For large datasets (10k+ players), a simple brute-force cosine similarity over ~30 attributes takes well under 100ms in a worker.
- Alternatively, the backend could pre-compute a similarity index at upload time for instant retrieval.
- A simpler heuristic: players in the same position group whose top-3 role suitabilities overlap — less accurate but near-zero compute cost.

**Effort estimate:** Low (heuristic) / Medium (full similarity computation).

---

## Cross-Cutting Notes

- Features 3 (Custom Rating Builder), 7 (Archetype Clustering), and 10 (Similar Players) all involve attribute-vector computation — there is an opportunity to share a common normalisation and weighting utility across all three.
- Features 5 (Multi-Dataset Comparison) and 6 (Squad Builder) would both benefit from the authentication system planned in roadmap #54 for dataset linking.
- Features 8 (PDF Scouting Report) and the AI summary it references are a natural pair with roadmap #59 (AI player descriptions).
- Feature 6 (Squad Builder) is a detailed elaboration of roadmap #60 (Create a team page).
