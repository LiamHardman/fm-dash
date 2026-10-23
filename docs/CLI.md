# FMDash CLI

Build from the Go module (Go 1.25 or newer):

```powershell
cd src/api
go build -o bin/fmdash.exe .
./bin/fmdash.exe help
```

On Linux/macOS use `go build -o bin/fmdash .`. The binary embeds its calculation defaults, formation layouts and supplied weight profiles; it works from other directories without a frontend build. Existing no-argument server startup is preserved. `fmdash serve` also starts the server. Every other command exits after its work completes.

## Improve a parser without a server

```text
fmdash parse preflight players.html
fmdash parse players players.html --strict --output before.json
fmdash parse validate --manifest testdata/cli/corpus.json
fmdash parse bench --manifest testdata/cli/corpus.json --iterations 10 --output baseline.json

# Rebuild after a parser change.
fmdash parse players players.html --strict --output after.json
fmdash parse diff --baseline before.json --candidate after.json
fmdash parse validate --manifest testdata/cli/corpus.json
fmdash parse bench --manifest testdata/cli/corpus.json --iterations 10 --baseline baseline.json --max-regression 5
```

The example manifest is relative to `src/api`. Fixture and golden paths inside a manifest resolve relative to that manifest; absolute paths are also supported. A baseline is never updated automatically. Use `--overwrite` to deliberately replace an output artifact.

CSV inputs use semicolons and support quoted delimiters, quoted multiline values, UTF-8 BOMs and large records. HTML reads the first nonempty table and supports both `th` and `td` headers. Full CSV header names are normalized using the same aliases as uploads. The GUI upload path uses the same decoder and parsing service.

`--stage parsed` decodes and normalizes fields without calculating ratings. `enhanced` adds individual ratings and derived statistics. The default, `complete`, also normalizes Dataset-wide MBR and calculates percentiles. `rawAttributes` preserves the original attribute values; it is aligned with `players` and `sourceRows`. Source rows are one-based logical table records, with the header at row 1; quoted multiline CSV records count as one record.

Player `division` and `basedIn` retain the export's values at every stage and through storage. Team and league summaries derive club-wide metadata separately, so loan rows can belong to their parent club's summary without changing player records. Previously imported Datasets whose metadata was overwritten must be imported again from the original export to recover those values.

Full CSV performance headers (including Minutes and Pass Completion Percentage) are mapped to the same numeric statistics as HTML exports. Sprints, key headers and aerial attempts are included in statistics and percentiles. `Rating` maps to `Av Rat`; `Average Rating Club` is retained as a separate statistic, including when their values differ. Transfer ranges use the upper endpoint, and monetary amounts round to the nearest whole currency unit, with half units rounded away from zero.

`--workers 1` is useful for diagnosis. Output order follows source order at every worker count. `--input-format csv|html` overrides the filename; stdin (`-`) requires this flag. `--max-bytes` bounds input size (default 500 MiB), and `--timeout` bounds command execution (default two minutes). Cancellation joins parser workers before returning. A supplied blocking stdin reader must itself unblock for the process to finish cancellation.

Preflight checks the complete FM Dash Search View column set and returns exit 3 for missing columns. Parsing permits narrower exports, but diagnoses missing Name/Position/Age columns, invalid required fields, cell-count mismatches and rejected rows. Permissive mode returns usable players with `status: partial` when there are errors. `--strict` turns these outcomes into exit 3. Missing/duplicate UIDs, currency fallback and mixed currencies are warnings; use source rows for ambiguous player identity. Mixed currencies are not converted. Repeated `Nat` columns are supported because FM uses that header for both nationality and Natural Fitness.

Native `.fm` inspection is separate:

```text
fmdash parse fm example.fm --output strings.json
```

This validates the header, decompresses a bounded prefix and returns plausible length-prefixed strings from the first 2 MiB. Offsets refer to the decompressed archive. `scannedBytes`, `scanLimit`, `truncated`, and string-length limits describe coverage. This command does not decode player records or create a Dataset. Experimental scanning parameters are not exposed yet.

## Accuracy manifests and benchmarks

The shipped corpus has independent expected counts and values for equivalent CSV/HTML exports. A case has this shape:

```json
{
  "version": 1,
  "cases": [{
    "name": "players",
    "path": "players.csv",
    "format": "csv",
    "stage": "complete",
    "expectedCount": 2,
    "expectedRejected": 0,
    "expectedErrors": 0,
    "expectedDiagnostics": [],
    "expected": [{"path": "/players/0/uid", "value": 101}]
  }]
}
```

`expectedCount` is required. `sha256` optionally pins fixture contents. `expected` paths are JSON pointers; `golden` optionally names a complete result artifact. Expected diagnostic codes, when supplied, must match exactly, including repetitions. Rejected rows and parse errors default to an expected count of zero. Large/private exports can be referenced locally without checking them into the repository.

`parse diff` matches unique positive UIDs and falls back to source-row identity for ambiguous UIDs. It excludes the envelope, timings and worker count; all remaining values are compared exactly, including missing versus null versus zero. It refuses input artifacts with different recorded input hashes. Numeric tolerances and comparison projections are not currently supported.

Benchmarks use one warm-up and ten measured iterations by default (`--warmup`, `--iterations`). Every iteration checks the corpus and must produce the same semantic output. Reports include read, decode, enhancement, Dataset calculation and serialization timings, total time, samples, median/p95, throughput, allocation bytes/count and Go heap bytes. Heap bytes are not process RSS. The total is a warm-process measurement; the operating-system file cache is uncontrolled. Run separate binary processes for cold experiments or different calculation configurations. Compilation and command output are outside the measurement.

Timing gates require a successful compatible baseline, matching fixture hashes, stages, worker counts, calculation configuration and reported runtime environment. They are opt-in. Correctness failures always fail evaluation, including benchmarks that produce too few rows. Small-fixture timings can be dominated by timer granularity; use representative inputs for optimization decisions. Ten samples provide only a coarse p95 estimate.

```text
fmdash parse bench --manifest corpus.json --cpu-profile cpu.pprof --heap-profile heap.pprof
go tool pprof cpu.pprof
```

## Local Dataset workflows

```text
fmdash datasets import players.html --workspace ./experiment --name "Season start"
fmdash datasets list --workspace ./experiment
fmdash players list --workspace ./experiment --dataset <id> --query '{"ageRange":{"min":18,"max":23},"minOverall":70,"sortField":"overall"}'
fmdash players show --workspace ./experiment --dataset <id> --id 101
fmdash players compare --workspace ./experiment --dataset <id> --id 101,102
fmdash players list --workspace ./experiment --dataset <id> --preset first-team-ready
fmdash datasets export --workspace ./experiment --dataset <id> --format csv --output players.csv
fmdash squad best-xi --workspace ./experiment --dataset <id> --name 4231_dm_am_wide --query '{"club":"Example FC"}'
```

Omit `--name` from `squad best-xi` to list the available formation IDs and layouts. `squad depth` returns the same three-level composition. Squad rules are tested against a fixture produced by the GUI algorithm, including scarcity, exact/fallback positions and role scoring.

Stateful commands default to `<OS user config directory>/fmdash`. Use an explicit workspace for experiments. Direct parsing/help/version/schema do not create a workspace. A workspace contains `datasets/<uuid>.json`, `profile/backup.json`, Managed club documents, calculation configuration and weight profiles. Dataset artifacts use FMDash's existing Player model and preserve raw attributes for recalculation. They have the configured Dataset retention period (30 days by default); expired Datasets are unavailable while durable Player references remain visible. Imports create fresh Dataset IDs. Dataset mutation and profile writes use an OS lock and atomic replacement; lock waits honor the deadline.

`datasets rename --name`, `datasets delete`, `ratings recalculate`, and `managed-club get/set/clear` operate on explicit `--dataset` IDs. A Managed club can be supplied with `--name` or `--body '{"club":"Example FC","division":"Test League"}'`. `datasets share --base-url` constructs a context URL; it does not upload a local Dataset to that backend.

Filters support names, clubs, positions, exact role names, nationality/division, personality/media handling, age/value/wage bounds, overall/CA/MBR, FIFA categories and individual attribute minima. For example, `minPAC` targets the derived Pace category and `minPac` targets the FM Pace attribute. Sort by `overall`, `age`, `value` or an FM attribute with `sortDir: asc|desc`. `--limit 0` returns all matches; otherwise the default is 100. Unsupported criteria fail explicitly. Explicit CLI numeric bounds apply even at zero; invalid numeric input fails instead of broadening the search.

Category exports accept the GUI options envelope, for example:

```text
fmdash datasets export --workspace ./experiment --dataset <id> --body '{"options":{"basicInfo":true,"fmStats":true,"roleRatings":true}}'
```

The switches are `basicInfo`, `fifahStats` (the GUI's existing spelling), `fmStats`, `roleRatings`, `performancePercentiles`, `contractInfo`, and `personalInfo`. Selected nested CSV values are JSON-encoded cells. Without switches, complete JSON/CSV exports use the existing backend exporters.

Additional local operations include `teams list/top`, `nations list/top-players --name`, `leagues list`, `divisions list`, `analysis save`, `analysis performance`, `recruitment bargains/free-agents/wonderkids/upgrades/recommend`, and `presets list`. Save analysis returns CA scaled to 99, outliers and role distribution; performance analysis returns underlying statistics, top performers, distributions and suggested thresholds. Wonderkids retain the GUI's top ten per age; free-agent detection includes empty clubs, dashes and unattached/free-agent labels. Upgrade requests use the existing backend request JSON. Recommendations take a body with `club`, `positions` and optional `minOverall`, `maxOverall`, `maxTransferValue`.

## Durable planning and settings

`shortlists list`, `saved-searches list`, and `progression projects` return a document and revision. Their corresponding `set` commands accept a complete GUI-compatible document through `--body @document.json` and require `--id <revision>`. Editing the document supports creation, updates, ordering, deletion, Recruitment status, Target budgets and scenario planning without adding separate permanent data models. A stale revision returns exit 7. Local revisions cover the whole Installation profile; backend revisions are obtained from backend ETags and must not be interchanged with local revisions.

```text
fmdash profile export --workspace ./experiment --output backup-result.json
fmdash profile validate --body @gui-backup.json
fmdash profile restore --workspace ./experiment --body @gui-backup.json --dry-run
fmdash profile restore --workspace ./experiment --body @gui-backup.json --replace
fmdash references resolve --workspace ./experiment --body '{"references":[{"datasetId":"<id>","playerUid":101}]}'
```

The `data` property of `profile export` is the GUI backup document. Restore merges by ID, keeping existing conflicts unless `--replace` is supplied. `--dry-run` returns the changes without replacing profile data. Shortlists, Saved searches and Progression projects contain references/metadata, never retained Dataset player records. Browser localStorage is not read automatically; export/import through supported documents explicitly.

`saved-searches run/explain --id <search-id> --dataset <id>` returns matches or current Match evidence. `recipes encode --body '<criteria JSON>'` and `recipes decode --query <recipe>` support GUI Search recipes without Dataset identity.

`progression analyze --body '{"datasetIds":["<a>","<b>"],"order":["<a>","<b>"]}'` compares snapshots. Omit `order` to infer it by average age; ambiguity fails with the candidate IDs rather than guessing. Snapshot records expose all values needed for trend and delta analysis. Repeated or incomplete explicit orders are rejected, and invalid/duplicate player UIDs are excluded from intersections.

`scenarios evaluate --id <scenario-id> --name <shortlist-id>` uses its Dataset and Managed club, manual Target budgets and live Player references. Missing targets and unresolved players are counted rather than presented as complete coverage.

`config show/validate/set` accepts `useScaledRatings` and `attributeWeights`. `weights list` includes supplied profiles and custom profiles; `weights set --body '{"id":"custom","name":"Custom","weights":{"PAC":{"Acc":10,"Pac":10}}}'` creates/updates/renames a custom profile. Use `weights activate --id <id>` (`default` for defaults) and `weights delete --id <id>`. Extract/import the returned profile objects to move custom profiles between workspaces. `--weights <path>` overrides attribute weights for one command. Explicit missing/invalid configuration fails. Recalculate existing Datasets after changing weights.

## Use an existing backend without the GUI

Provider-backed scouting, the Scout Assistant and server operational features use HTTP, with no browser dependency:

```text
fmdash datasets upload players.html --base-url http://localhost:8091
fmdash scout reports --dataset <backend-id> --base-url http://localhost:8091
fmdash scout report --dataset <backend-id> --body @report-request.json --header 'X-OpenAI-Api-Key:@OPENAI_API_KEY'
fmdash assistant ask --dataset <backend-id> --body @conversation.json --header 'X-OpenAI-Api-Key:@OPENAI_API_KEY'
fmdash api /api/players/<backend-id> --query 'position=MC&minAge=18'
fmdash api /api/shortlists
fmdash api /api/shortlists --method PUT --id '<etag>' --body @shortlists.json
```

`datasets upload` returns the backend's Dataset ID; a local workspace Dataset ID is not automatically available on that backend. `api /api/...` exposes existing backend query/body contracts, including full player details/percentiles, team and league analysis, exports, configuration, profile documents, legacy wishlists, progression and club mapping. For multipart endpoints use `--method POST --upload <path>` (`playerFile`, or `saveFile` on `/api/fm-save-import`).

Named backend operations also include `scout stars` (query `playerUid` and `position`), `recruitment who-to-sign`, `cache status/clear`, `debug player`, and `club-match resolve/overrides`. Use `api` with an explicit method for report reads/deletes or other variations. Headers are repeatable; `NAME:@ENVIRONMENT_VARIABLE` reads the value without placing it in shell arguments. The same `X-OpenAI-Base-URL` and `X-OpenAI-Model` headers as the GUI are supported. Credentials are not included in provenance. Responses include HTTP status, ETag and data; event-stream/text responses are preserved as strings. Redirects are not followed automatically.

These operations intentionally use the existing backend for its provider configuration and persisted reports. They require a running backend and any prerequisites the GUI requires, including a valid Managed club for the Scout Assistant. Local parser validation/benchmarking remains offline.

## Output and failures

JSON is the default: `schemaVersion`, `command`, `status`, `data`, `diagnostics`, `provenance`. `help`, `capabilities` and `schema` describe commands, flags and the envelope. `--format table` renders readable data. Parser JSONL emits a header, source-identified player records and a terminal summary; a missing successful summary is incomplete output. Other commands emit one JSON envelope per line.

Stdout contains results; stderr contains logs. `--quiet` and `--verbose` control diagnostic logging. File outputs commit atomically and require `--overwrite` to replace an existing path. Failed parsing/workflow commands leave the requested artifact untouched and emit their failure envelope to stdout. Failed validation/diff/benchmark commands can still write their complete evaluation report.

| Exit | Meaning |
| --- | --- |
| 0 | Success, including explicitly reported permissive partial parsing |
| 1 | Internal failure |
| 2 | Usage/configuration error |
| 3 | Invalid input or strict rejection |
| 4 | Accuracy or performance gate failed |
| 5 | Unavailable Dataset, reference or capability |
| 6 | Output/storage/network/provider failure |
| 7 | Revision conflict |
| 124 | Deadline exceeded |
| 130 | Interrupted |

The implementation coverage and test pointers are in [the parity matrix](../.scratch/parsing-cli/parity.md). Run `go test ./...` from `src/api` and `npm run test:run` from the repository root. The CLI's Go tests include worker determinism, corruption/rejections, benchmark gates, output safety, workspace persistence/concurrent mutation, revision conflicts and mocked backend transfers. GUI tests detect drift in embedded formation/weight assets and the squad parity fixture.
