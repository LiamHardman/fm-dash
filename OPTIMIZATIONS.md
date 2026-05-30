# Data Transfer Optimizations

Six concrete changes ranked by ease vs. impact. The first four are low-effort and together should cut typical payload sizes by 40–60% and eliminate redundant round-trips.

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
