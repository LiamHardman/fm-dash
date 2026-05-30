# Data Transfer Optimizations

Six concrete changes ranked by ease vs. impact. The first four are low-effort and together should cut typical payload sizes by 40–60% and eliminate redundant round-trips.

---

## 1. Remove duplicate fields from the Player struct

**File**: `src/api/types.go:51–75`

### Problem

The `Player` struct carries explicit uppercase **and** lowercase versions of 13 fields:

```go
PAC int `json:"PAC"`
Pac int `json:"pac,omitempty"`  // duplicate

Overall      int `json:"Overall"`
OverallLower int `json:"overall"`  // duplicate

CA int `json:"CA"`
Ca int `json:"ca,omitempty"`  // duplicate

// ... TotalStats/totalStats, MBR/mbr, SHO/sho, PAS/pas,
//     DRI/dri, DEF/def, PHY/phy, GK/gk, DIV/div,
//     HAN/han, REF/ref, KIC/kic, SPD/spd, POS/pos
```

Every player object sent over the wire carries both versions. For a 10,000-player dataset that's roughly 130,000 redundant JSON field emissions per request.

### What the frontend uses

The player **table** (`src/components/player-table/PlayerTableColumns.js`) uses lowercase field names (`pac`, `overall`, `totalStats`, etc.). The uppercase variants appear only in a few column-definition names (display labels), not as data accessors.

### Fix

1. Keep only the **lowercase** json tags on each field — they match what the table and store already read.
2. Remove the duplicate `Overall`/`OverallLower` pair; keep `Overall int \`json:"overall"\``.
3. Search for uppercase field access in the frontend (`player.PAC`, `player.Overall`, etc.) and normalise to lowercase.
4. Update any CSV export or protobuf mappings that reference the old uppercase keys.

```go
// Before
Overall      int `json:"Overall"`
OverallLower int `json:"overall"`

// After
Overall int `json:"overall"`
```

**Expected saving**: ~10–15% payload reduction on top-level fields alone; more significant for GK players where all GK variants are non-zero.

---

## 2. Stop actively disabling browser caching

**File**: `src/services/playerService.js:38–45`

### Problem

Every player data request is sent with headers that tell the browser and any intermediate proxy to never cache the response:

```js
cache: 'no-store',
headers: {
  'Cache-Control': 'no-cache',
  Pragma: 'no-cache',
}
```

This means that if a user switches from the Squad view to the Nations page and back, the entire dataset (potentially several MB of JSON) is re-fetched from the server. There is no stale data risk here because dataset IDs are content-addressed — the data for a given `datasetId` never changes after upload.

### Fix

Remove the cache-busting headers from `fetchJsonResponse`. Optionally, tell the server to set a short-lived `Cache-Control` header so the browser can use its cache for the duration of a session:

```js
// Before
const response = await fetch(url, {
  ...options,
  cache: 'no-store',
  headers: {
    ...(options.headers || {}),
    Accept: 'application/json',
    'Cache-Control': 'no-cache',
    Pragma: 'no-cache',
  },
})

// After
const response = await fetch(url, {
  ...options,
  headers: {
    ...(options.headers || {}),
    Accept: 'application/json',
  },
})
```

On the server side, add a `Cache-Control` header to player-data responses:

```go
w.Header().Set("Cache-Control", "private, max-age=300")
```

`private` means the browser can cache it but CDNs won't share it between users. `max-age=300` (5 minutes) is conservative and safe — datasets don't mutate after upload.

**Expected saving**: Eliminates full re-fetches on in-session navigation. On a 5MB dataset, each navigation back to the squad page currently costs a full round-trip; after this change it costs nothing.

---

## 3. Lower the streaming threshold

**File**: `src/api/streaming_serialization.go:24`

### Problem

```go
const maxStreamingPlayers = 100000 // Don't stream for small datasets
```

Real FM exports are typically 5,000–50,000 players. Because this threshold is 100K, streaming is **never** activated in practice. The entire response is serialised into a single large buffer in memory before the first byte reaches the client. The client's rendering pipeline sits idle waiting for the full payload.

### Fix

Lower the threshold so streaming activates for any dataset large enough to matter:

```go
const maxStreamingPlayers = 2000
```

With streaming active, the client receives and can begin rendering the first chunk of players within ~100ms while the server continues sending the rest. Memory pressure on the server also drops because it never needs to hold the full serialised payload in RAM.

If 2,000 is too aggressive (e.g. small datasets where chunking overhead outweighs the benefit), 5,000 is a safe middle ground. The chunking infrastructure is already written and tested — this is a one-line change.

**Expected saving**: First-meaningful-paint time for large datasets improves significantly; server peak memory per request drops proportionally to dataset size.

---

## 4. Exclude `PerformancePercentiles` from list responses

**File**: `src/api/types.go:33`, `src/api/handlers.go`

### Problem

```go
PerformancePercentiles map[string]map[string]float64 `json:"performancePercentiles"`
```

This is a nested map (position group → stat name → float64 percentile) included in every player in every list response. It is only rendered in one place: the **detail dialog** (`src/components/PlayerDetailDialog.vue:295–410`) and `PerformanceAnalysisCard.vue`. The player list table never reads it.

For a player with 5 position groups and 20 stats each, this adds ~100 float64 values per player. Across 10,000 players that is 1,000,000 floats that the client silently discards when rendering the table.

### Fix

**Option A — Query param (recommended)**

Add an `?include=percentiles` parameter. The handler omits percentiles by default and only populates them when the param is present:

```go
// In the player list handler
includePercentiles := r.URL.Query().Get("include") == "percentiles"
if !includePercentiles {
    for i := range players {
        players[i].PerformancePercentiles = nil
    }
}
```

The frontend `PlayerDetailDialog` already has its own percentile-fetching logic (`src/composables/usePercentileRetry.js`) — confirm it is the one requesting with `?include=percentiles` and that the list endpoint is not.

**Option B — Separate endpoint**

Move percentile data to `GET /api/players/{datasetId}/{uid}/percentiles`. The detail dialog fetches it on open. This is cleaner long-term but requires a new route and a store update.

**Expected saving**: Rough estimate 20–40% payload reduction on list responses for datasets with rich performance data. The saving is proportional to how many players have performance stats populated.

---

## 5. Enable Protobuf by default

**File**: `src/api/storage_init.go:81`, deployment config / `docker-compose.yml`

### Problem

Protobuf serialisation is fully implemented — the server supports content negotiation (`src/api/content_negotiation.go`), the frontend has a protobuf client (`src/utils/protobufClient.js`, `src/composables/useProtobufApi.js`), and the storage layer has a protobuf wrapper. Despite this, it is gated behind an environment variable that defaults to `false`:

```go
case "false", "0", "no", "off", "":
    config.UseProtobuf = false
```

JSON is human-readable but verbose. Protobuf binary encoding is typically 40–60% smaller for the same data and faster to serialise/deserialise.

### Fix

Set `USE_PROTOBUF=true` in your deployment environment. For local development, add it to your `.env` or `docker-compose.yml`:

```yaml
environment:
  - USE_PROTOBUF=true
```

Before flipping this on in production, verify:

1. The protobuf `Player` schema in `src/api/proto/` is in sync with the Go `Player` struct in `types.go` — any field added to the struct but not the `.proto` file will be silently dropped in protobuf responses.
2. The frontend protobuf client handles all edge cases (nil maps, empty slices) identically to the JSON path. Run through the main flows manually or check the integration tests in `src/api/player_data_handler_protobuf_integration_test.go`.
3. The content negotiation fallback to JSON still works when the client does not send `Accept: application/x-protobuf`.

**Expected saving**: 40–60% reduction in transfer size across all endpoints once the frontend is sending the protobuf `Accept` header. Combined with gzip (already active), the effective saving stacks.

---

## 6. Remove `attributes` (string map) from list responses

**File**: `src/api/types.go:30`

### Problem

```go
Attributes        map[string]string `json:"attributes"`
NumericAttributes map[string]int    `json:"numericAttributes"`
```

`attributes` is the raw string version of FM attribute values (e.g. `{"Finishing": "14", "Dribbling": "17"}`). `numericAttributes` is the same data parsed to integers. Both are sent in every player object.

The string map is used in two places:
- `PlayerDetailDialog.vue` — reads `player.attributes[attrKey]` to colour-code attribute ratings.
- `csvExport.js` — enumerates attribute keys when building a CSV.

The numeric map is used for calculations and sorting in the store.

They contain the same keys; one is strictly redundant on the wire.

### Fix

**Drop `attributes` from the JSON response** and derive it in the frontend from `numericAttributes`:

```js
// In playerStore.js normalization, after receiving numericAttributes:
if (!player.attributes && player.numericAttributes) {
  player.attributes = Object.fromEntries(
    Object.entries(player.numericAttributes).map(([k, v]) => [k, String(v)])
  )
}
```

Then on the server, set the json tag to `-` so it is never serialised:

```go
Attributes map[string]string `json:"-"`
```

Keep `Attributes` populated in memory (it is used during HTML parsing), just don't send it over the wire.

**Caveat**: Confirm that the string map and numeric map always have identical keys. If `attributes` contains keys not present in `numericAttributes` (e.g. text-only fields), those would need to be handled separately or moved to a dedicated text-attributes field.

**Expected saving**: Roughly equivalent to removing `numericAttributes` (~20–30 fields per player). For a 10,000-player dataset this is in the range of 1–3MB of JSON depending on attribute coverage.

---

## Combined impact estimate

Applying all six changes to a typical 10,000-player dataset:

| Change | Estimated saving |
|---|---|
| Remove duplicate fields (#1) | ~8–12% |
| Browser caching (#2) | Eliminates repeat fetches entirely |
| Streaming threshold (#3) | Reduces TTFB, not raw size |
| Exclude percentiles from lists (#4) | ~20–35% |
| Protobuf encoding (#5) | ~40–55% of remaining size |
| Remove string attributes map (#6) | ~8–15% |

Changes #1, #4, and #5 stack multiplicatively. The realistic outcome for a cold load of a 50,000-player dataset is a reduction from ~80–120MB of JSON to ~15–25MB of protobuf, with the client able to start rendering after the first streaming chunk rather than waiting for the full payload.
