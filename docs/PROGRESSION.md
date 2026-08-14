# Progression — Design Spec

Status: **Ready for review**
Working name: "Progression" (previously "Development Over Time")

## 1. Overview & Goals

**Progression** is a new top-level section (route `/progression`, own nav item) where a
user uploads 2+ FM export files representing the same save at different points in time.
The app determines chronological order automatically, computes the set of players present
in *every* upload (matched by `Player.UID`), calculates deltas between each player's first
and last snapshot for any numeric stat, and displays a filterable/sortable table with a
trend view across all snapshots.

**Non-goals:**
- No persistence of the "comparison" itself as a shareable entity — it's a session-scoped
  view composed from ordinary datasets.
- No cross-save player matching (UID is only stable within one continuous save).
- No editing/annotation features.

## 2. Backend Data Model & Flow

No new persistent entity type — Progression is computed on-demand from existing
`DatasetData` records. New file `src/api/progression_handler.go`:

1. **Upload** — frontend uploads each file via the existing `/api/upload` endpoint
   (unchanged), collecting the resulting `datasetID`s client-side as the user adds
   snapshots.
2. **`POST /api/progression/analyze`** — body: `{ "datasetIds": ["uuid1", "uuid2", ...] }`.
   Handler:
   - Loads each dataset via existing `GetPlayerData(datasetID)`.
   - Computes the N-way intersection of players by `UID` (build a `map[int64]int` counting
     appearances across datasets; keep UIDs with count == len(datasetIds)).
   - Determines order: for each dataset, compute mean `Age` over the intersected UID set;
     sort dataset IDs ascending by mean age. If two or more means are equal (tie), return
     an `orderAmbiguous: true` flag plus the tied dataset IDs instead of guessing, so the
     frontend can prompt for manual ordering.
   - Once ordered, builds a `ProgressionPlayer` per shared UID:
     `{ UID, Name, snapshots: []PlayerSnapshot }` where `PlayerSnapshot` is a trimmed
     per-dataset view `{ datasetId, Overall, Value, Age, Attributes map[string]int, ... }`
     pulled straight from each dataset's already-parsed `Player` struct — no re-parsing.
   - Returns `{ order: [datasetId...], players: []ProgressionPlayer }`.
3. Filtering and sorting (see section 4) happen server-side in the same handler via query
   params/body, mirroring the existing `playerDataHandler` filter block — reused, not
   reimplemented, by extracting the current inline filter logic in `handlers.go` into a
   shared helper `applyPlayerFilters(players []Player, filters FilterParams) []Player` that
   both `playerDataHandler` and the new progression handler call against each player's
   **latest** snapshot.

This keeps Progression a thin composition layer: it never duplicates upload/parsing/storage
code, and the only genuinely new logic is intersection + ordering + delta computation.

## 3. Frontend UI & UX

- **Route/nav**: `/progression`, new sidebar entry "Progression" (Vue Router + Quasar,
  matching existing nav pattern).
- **Page**: `src/pages/ProgressionPage.vue`, with a Pinia store
  `src/stores/progressionStore.js` and `src/services/progressionService.js` (mirrors the
  wishlist trio).
- **Upload state** (before analysis):
  - A list of upload slots, each just the existing upload widget/component reused,
    "+ Add another snapshot" button to add more (no upper bound).
  - Each slot shows filename + a small status (uploading/parsed/error) once its
    `/api/upload` call resolves and yields a `datasetID`.
  - "Analyze" button enabled once ≥2 slots have a `datasetID`; calls
    `POST /api/progression/analyze`.
  - If the response has `orderAmbiguous: true`, show a drag-to-reorder list of the
    ambiguous datasets (by filename) with a "Confirm order" button before re-requesting
    analysis with an explicit `order` array override.
- **Results state** (after analysis):
  - A table of common players: Name, Position, current Club, then dynamic delta columns.
    Default visible: Overall Δ, Value Δ.
  - A "Sort by change in…" dropdown (the generic numeric-field picker) plus asc/desc toggle
    — replaces bespoke "most/least improved" buttons with one generic control seeded with
    Overall selected by default.
  - Existing filter panel (reused component from `DatasetPage.vue`/`usePlayerFilters.js`)
    docked the same way it is today, applied against latest-snapshot values.
  - Row expansion or a hover/click reveals the trend view: a small sparkline per numeric
    stat across all snapshots (reusing whatever charting primitive, if any, already exists
    in the codebase — otherwise a minimal inline SVG sparkline, no new charting library).

## 4. Filtering & Sorting Integration

- **Filters**: no new filter logic. The extracted `applyPlayerFilters` helper (from
  section 2) is called with each `ProgressionPlayer`'s *latest* snapshot's `Player`-shaped
  data, and only players that pass are kept in the response — this is exactly today's
  single-dataset filter behavior, just fed a pre-intersected list instead of a full
  dataset.
- **Generic delta sort**:
  - Backend: `sortField` query param (e.g. `Overall`, `Value`, `Attributes.Pace`) +
    `sortDir` (`asc`/`desc`). For each `ProgressionPlayer`, delta =
    `last.snapshots[N-1].Field - first.snapshots[0].Field`, computed via reflection/a small
    numeric-field accessor map (not per-field switch statements) so any current or future
    numeric `Player` field works without backend changes.
  - Frontend dropdown is populated from a small static list of "known interesting fields"
    (Overall, Value, PA, Wage, Age, + all attribute keys) rather than introspecting the
    struct client-side, since attribute names are already enumerable from the existing
    player data the frontend receives.
- **Interaction with existing division/top5 filters**: unaffected — those filters already
  operate on a single `Player`'s fields (division, comparison flags), so they apply to the
  latest-snapshot view the same way.

## 5. Error Handling & Edge Cases

- **Empty intersection** (no player UID present in all datasets): return `players: []`
  with a distinct `emptyIntersection: true` flag so the frontend can show a clear message
  ("No players found in every uploaded save — check these are from the same game") rather
  than an empty-looking table.
- **Fewer than 2 valid datasets** (e.g. an upload failed): "Analyze" stays disabled
  client-side; server also validates `len(datasetIds) >= 2` and 400s otherwise.
- **Different saves uploaded by mistake**: no explicit "same save" detection is planned
  (out of scope) — an empty or tiny intersection is the natural signal to the user
  something's wrong, covered by the empty-intersection message above.
- **Dataset expiry/cleanup**: existing datasets can be garbage-collected by
  `CleanupOldDatasets`; if a `datasetID` passed to `/api/progression/analyze` no longer
  resolves, return 404 for that ID specifically so the frontend can prompt "please
  re-upload that snapshot" rather than failing silently.
- **Large N or large rosters**: intersection is a single O(total players) pass over a map;
  no pagination planned initially given typical roster sizes (hundreds, not millions) —
  revisit only if real usage shows a problem.
- **Duplicate file upload** (same file uploaded twice as two "different" snapshots):
  existing duplicate-hash detection in the upload handler is unrelated/unaffected — two
  identical files would just produce two datasets with a mean-age tie between them,
  correctly triggering the manual-order-confirmation flow (harmless, if slightly
  meaningless, comparison).

## 6. Testing Plan

- **Backend unit tests** (`src/api/progression_handler_test.go`):
  - Intersection: overlapping/disjoint/fully-identical UID sets across 2 and 3+ datasets.
  - Ordering: distinct mean ages sort correctly; tied means produce `orderAmbiguous`.
  - Delta calc: correct first-vs-last deltas for Overall/Value/an attribute, including
    negative deltas (regression) and zero deltas.
  - Filter reuse: `applyPlayerFilters` produces identical results whether called from
    `playerDataHandler` or the progression handler, given equivalent input
    (regression-guards the extraction).
  - Edge cases from section 5: empty intersection, <2 datasets, missing/expired datasetID.
- **Frontend**: component/unit tests for `progressionStore.js` (state transitions:
  uploading → analyzed → ambiguous-order → confirmed) and the generic sort-field dropdown;
  existing filter-panel tests should already cover filter behavior since the component is
  reused as-is.
- **Manual/E2E pass**: upload 2-3 real FM exports from the same save via the running dev
  app, confirm ordering, deltas, and filters behave as expected before calling the feature
  done (per this project's standard of verifying UI changes in-browser).
