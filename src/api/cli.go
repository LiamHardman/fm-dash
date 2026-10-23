package main

import (
	"context"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

//go:embed public/attribute_weights.json public/role_specific_overall_weights.json
var cliDefaults embed.FS

var cliInvocationMutex sync.Mutex

var cliVersion = "dev"

type cliError struct {
	code int
	err  error
}

func (e *cliError) Error() string { return e.err.Error() }
func (e *cliError) Unwrap() error { return e.err }
func cliFail(code int, format string, values ...any) error {
	return &cliError{code, fmt.Errorf(format, values...)}
}

type cliEnvelope struct {
	SchemaVersion int               `json:"schemaVersion"`
	Command       string            `json:"command"`
	Status        string            `json:"status"`
	Data          any               `json:"data"`
	Diagnostics   []ParseDiagnostic `json:"diagnostics"`
	Provenance    map[string]any    `json:"provenance"`
}

type cliOptions struct {
	workspace, format, output                                string
	inputFormat, stage, weights                              string
	dataset, id, name, query, body                           string
	baseURL, method, upload                                  string
	manifest, baseline, candidate                            string
	workers, iterations, warmup, limit, offset               int
	maxBytes                                                 int64
	timeout                                                  time.Duration
	strict, overwrite, quiet, verbose, replace, dryRun, help bool
	maxRegression                                            float64
	cpuProfile, heapProfile                                  string
	headers                                                  cliHeaders
	preset                                                   string
}

var cliCommands = map[string]string{
	"help": "Describe commands and common flags", "version": "Build and runtime information",
	"capabilities": "Machine-readable command inventory", "schema": "Describe flags and envelope",
	"parse preflight": "Validate CSV/HTML export headers", "parse players": "Parse CSV/HTML without a server",
	"parse fm": "Inspect bounded strings from a native .fm archive", "parse validate": "Check an annotated corpus manifest",
	"parse diff": "Compare semantic parser output", "parse bench": "Benchmark a validated corpus",
	"datasets import": "Persist a parsed Dataset in a workspace", "datasets list": "List workspace Datasets",
	"datasets upload": "Upload an export to an existing backend (returns a backend Dataset ID)",
	"datasets show":   "Read a workspace Dataset", "datasets delete": "Delete a workspace Dataset",
	"datasets rename": "Change a Dataset label", "datasets export": "Export Dataset players as JSON or CSV",
	"datasets share": "Return a Dataset context URL for a configured backend",
	"players list":   "Filter and sort Dataset players", "players show": "Inspect one player by UID",
	"players compare": "Compare explicit comma-separated player UIDs", "players search": "Search player names and clubs",
	"teams list": "Team summaries", "teams top": "Rank team summaries", "nations list": "Nation summaries",
	"nations top-players": "Select a nation's squad", "leagues list": "League summaries", "divisions list": "Division names",
	"managed-club get": "Read the Managed club", "managed-club set": "Set the Managed club", "managed-club clear": "Clear the Managed club",
	"config show": "Show resolved calculation configuration", "config set": "Validate and persist workspace configuration",
	"config validate": "Validate calculation configuration", "ratings recalculate": "Recalculate a workspace Dataset",
	"recipes encode": "Encode Saved search criteria", "recipes decode": "Decode a Search recipe",
	"profile export": "Export manager-authored documents", "profile validate": "Validate a GUI profile backup",
	"profile restore": "Preview or restore a profile backup", "references resolve": "Resolve durable Player references",
	"shortlists list": "Read Shortlists and transfer planning", "shortlists set": "Revision-checked Shortlist document replacement",
	"saved-searches list": "Read Saved searches", "saved-searches set": "Revision-checked Saved search document replacement",
	"progression projects": "Read Progression projects", "progression set": "Revision-checked Progression project document replacement",
	"progression analyze":  "Analyze explicitly ordered Dataset snapshots",
	"recruitment bargains": "Find affordable candidates", "recruitment free-agents": "Find unattached players",
	"recruitment wonderkids": "Find young candidates", "recruitment upgrades": "Find role-specific upgrades",
	"recruitment recommend": "Find position-specific signings relative to the current squad",
	"squad best-xi":         "Select formation roles", "squad depth": "Rank depth for formation roles",
	"analysis save": "Return rating outliers and role distribution", "analysis performance": "Return numeric performance statistics",
	"status": "Workspace/runtime status", "api": "Call an existing GUI backend without a browser",
	"presets list": "List GUI Starter presets", "saved-searches run": "Run stored discovery criteria", "saved-searches explain": "Describe criteria and current Match evidence",
	"scenarios evaluate": "Evaluate a Shortlist recruitment scenario",
	"weights list":       "List custom and supplied weight profiles", "weights set": "Create or replace a custom weight profile",
	"weights activate": "Activate a weight profile", "weights delete": "Delete a custom weight profile",
	"scout report": "Generate or manage a report through the backend", "scout stars": "Read squad/division star ratings through the backend",
	"scout reports": "Read the backend scouting book", "assistant ask": "Ask the backend Scout Assistant",
	"recruitment who-to-sign": "Request backend squad recruitment advice",
	"cache status":            "Read backend cache status", "cache clear": "Clear a named backend cache",
	"debug player": "Inspect backend player data", "club-match resolve": "Resolve a club identity through the backend",
	"club-match overrides": "Read or change backend identity overrides",
}

func cliFlagSet(o *cliOptions) *flag.FlagSet {
	f := flag.NewFlagSet("fmdash", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	f.BoolVar(&o.help, "help", false, "Show command help")
	f.BoolVar(&o.help, "h", false, "Show command help")
	f.Var(&o.headers, "header", "API header NAME:VALUE; repeatable; use @ENV_NAME as value for secrets")
	f.StringVar(&o.preset, "preset", "", "Starter preset applied before explicit --query criteria")
	f.StringVar(&o.upload, "upload", "", "File to upload as multipart data for api commands")
	f.StringVar(&o.workspace, "workspace", "", "State directory (default: OS user config directory/fmdash)")
	f.StringVar(&o.format, "format", "json", "Output format: json, jsonl, table; datasets export also supports csv")
	f.StringVar(&o.output, "output", "", "Write atomically to a path instead of stdout")
	f.BoolVar(&o.overwrite, "overwrite", false, "Replace an existing output artifact")
	f.DurationVar(&o.timeout, "timeout", 2*time.Minute, "Command deadline")
	f.BoolVar(&o.quiet, "quiet", false, "Suppress informational logs")
	f.BoolVar(&o.verbose, "verbose", false, "Include debug logs on stderr")
	f.StringVar(&o.inputFormat, "input-format", "", "Input format override: csv, html, fm")
	f.StringVar(&o.stage, "stage", "complete", "Calculation stage: parsed, enhanced, complete")
	f.IntVar(&o.workers, "workers", min(runtime.NumCPU(), 256), "Parser worker count (1-256)")
	f.Int64Var(&o.maxBytes, "max-bytes", 500<<20, "Maximum input bytes")
	f.BoolVar(&o.strict, "strict", false, "Fail on any parse error or rejected row")
	f.StringVar(&o.weights, "weights", "", "Explicit attribute weights JSON file")
	f.StringVar(&o.dataset, "dataset", "", "Dataset identifier (required for Dataset-scoped commands)")
	f.StringVar(&o.id, "id", "", "Player UID(s) or document revision for writes")
	f.StringVar(&o.name, "name", "", "Label, club, nation or role")
	f.StringVar(&o.query, "query", "", "JSON filter object, search text, or API query string")
	f.StringVar(&o.body, "body", "", "Request JSON or @path (use @- for stdin)")
	f.StringVar(&o.baseURL, "base-url", "http://localhost:8091", "Backend URL for api commands")
	f.StringVar(&o.method, "method", "GET", "HTTP method for api")
	f.StringVar(&o.manifest, "manifest", "", "Corpus manifest path")
	f.StringVar(&o.baseline, "baseline", "", "Baseline result or benchmark report")
	f.StringVar(&o.candidate, "candidate", "", "Candidate parser result")
	f.IntVar(&o.iterations, "iterations", 10, "Measured benchmark iterations")
	f.IntVar(&o.warmup, "warmup", 1, "Benchmark warm-up iterations")
	f.IntVar(&o.limit, "limit", 100, "Result limit (0 means all)")
	f.IntVar(&o.offset, "offset", 0, "Result offset")
	f.Float64Var(&o.maxRegression, "max-regression", 5, "Maximum benchmark regression in percent")
	f.StringVar(&o.cpuProfile, "cpu-profile", "", "CPU profile artifact path")
	f.StringVar(&o.heapProfile, "heap-profile", "", "Heap profile artifact path")
	f.BoolVar(&o.replace, "replace", false, "Replace conflicting profile records during restore")
	f.BoolVar(&o.dryRun, "dry-run", false, "Preview a mutation without writing")
	return f
}

// splitCLI permits flags before or after positionals, including input filenames.
func splitCLI(args []string, flags *flag.FlagSet) ([]string, []string, error) {
	var options, positional []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}
		if a == "-" || !strings.HasPrefix(a, "-") {
			positional = append(positional, a)
			continue
		}
		key := strings.TrimLeft(strings.SplitN(a, "=", 2)[0], "-")
		f := flags.Lookup(key)
		if f == nil {
			return nil, nil, cliFail(2, "unknown flag %s", a)
		}
		options = append(options, a)
		if !strings.Contains(a, "=") {
			if b, ok := f.Value.(interface{ IsBoolFlag() bool }); ok && b.IsBoolFlag() {
				continue
			}
			i++
			if i == len(args) {
				return nil, nil, cliFail(2, "missing value for %s", a)
			}
			options = append(options, args[i])
		}
	}
	return options, positional, nil
}

func cliBuildInfo() map[string]any {
	info := map[string]any{"go": runtime.Version(), "os": runtime.GOOS, "arch": runtime.GOARCH, "cpus": runtime.NumCPU(), "gomaxprocs": runtime.GOMAXPROCS(0)}
	info["version"] = cliVersion
	info["cpuIdentifier"] = os.Getenv("PROCESSOR_IDENTIFIER")
	info["gogc"] = os.Getenv("GOGC")
	info["gomemlimit"] = os.Getenv("GOMEMLIMIT")
	if b, ok := debug.ReadBuildInfo(); ok {
		for _, s := range b.Settings {
			if strings.HasPrefix(s.Key, "vcs") {
				info[s.Key] = s.Value
			}
		}
	}
	return info
}

func cliInitializeConfig(o cliOptions) error {
	// Each invocation starts from identical shipped defaults, even in embedded
	// tests that run multiple commands in one process.
	attr, _ := cliDefaults.ReadFile("public/attribute_weights.json")
	role, _ := cliDefaults.ReadFile("public/role_specific_overall_weights.json")
	var attrs, roles map[string]map[string]int
	if err := json.Unmarshal(attr, &attrs); err != nil {
		return err
	}
	if err := json.Unmarshal(role, &roles); err != nil {
		return err
	}
	muAttributeWeights.Lock()
	attributeWeights = attrs
	muAttributeWeights.Unlock()
	muRoleSpecificOverallWeights.Lock()
	roleSpecificOverallWeights = roles
	muRoleSpecificOverallWeights.Unlock()
	precomputeRoleWeights()
	configInitOnce.Do(func() { configInitialized = true })
	SetUseScaledRatings(true)
	w, err := cliOpenWorkspace(o.workspace)
	if err != nil {
		return err
	}
	configuration, err := os.ReadFile(filepath.Join(w.dir, "config.json"))
	if err != nil && !os.IsNotExist(err) {
		return cliFail(2, "workspace configuration: %v", err)
	}
	if err == nil {
		var saved struct {
			AttributeWeights map[string]map[string]int `json:"attributeWeights"`
			UseScaledRatings *bool                     `json:"useScaledRatings"`
		}
		if err = cliDecodeJSON(configuration, &saved); err != nil {
			return err
		}
		if saved.AttributeWeights != nil {
			if err = SetAttributeWeights(saved.AttributeWeights); err != nil {
				return cliFail(2, "workspace configuration: %v", err)
			}
		}
		if saved.UseScaledRatings != nil {
			SetUseScaledRatings(*saved.UseScaledRatings)
		}
	}
	if o.weights != "" {
		data, err := os.ReadFile(o.weights)
		if err != nil {
			return cliFail(2, "weights: %v", err)
		}
		var weights map[string]map[string]int
		if err = json.Unmarshal(data, &weights); err != nil {
			return cliFail(2, "weights: %v", err)
		}
		if err = SetAttributeWeights(weights); err != nil {
			return cliFail(2, "weights: %v", err)
		}
	}
	return nil
}

func cliReadInput(ctx context.Context, path string, stdin io.Reader, maxBytes int64) ([]byte, error) {
	r := stdin
	if path != "-" {
		f, err := os.Open(path)
		if err != nil {
			return nil, cliFail(3, "input: %v", err)
		}
		defer func() { _ = f.Close() }()
		r = f
	}
	if maxBytes < 1 || maxBytes > 1<<40 {
		return nil, cliFail(2, "max-bytes must be between 1 and 1 TiB")
	}
	data, err := io.ReadAll(io.LimitReader(contextReader{ctx, r}, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, cliFail(3, "input exceeds %d bytes", maxBytes)
	}
	return data, nil
}

func cliHash(data []byte) string { hash := sha256.Sum256(data); return hex.EncodeToString(hash[:]) }

func cliWriteFile(path string, data []byte, overwrite bool) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".fmdash-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer func() { _ = os.Remove(tmp) }()
	if _, err = f.Write(data); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if overwrite {
		return os.Rename(tmp, path)
	}
	// Linking commits only if the destination does not exist (including races).
	return os.Link(tmp, path)
}

// RunCLI never starts a server and is testable with explicit streams.
func RunCLI(parent context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	cliInvocationMutex.Lock()
	defer cliInvocationMutex.Unlock()
	var o cliOptions
	f := cliFlagSet(&o)
	options, positional, err := splitCLI(args, f)
	if err == nil {
		err = f.Parse(options)
		if err != nil {
			err = cliFail(2, "%v", err)
		}
	}
	command := "help"
	var rest []string
	if len(positional) > 0 {
		command = positional[0]
		rest = positional[1:]
		if len(rest) > 0 {
			if _, ok := cliCommands[command+" "+rest[0]]; ok {
				command += " " + rest[0]
				rest = rest[1:]
			}
		}
	}
	if o.help {
		command = "help"
		rest = nil
	}
	envelope := cliEnvelope{1, command, "ok", nil, []ParseDiagnostic{}, cliBuildInfo()}
	if err == nil && o.timeout <= 0 {
		err = cliFail(2, "timeout must be positive")
	}
	if err == nil && (o.maxBytes < 1 || o.maxBytes > 1<<40 || o.workers < 1 || o.workers > 256 || o.limit < 0 || o.offset < 0) {
		err = cliFail(2, "invalid bounds: max-bytes 1..1TiB, workers 1..256, limit/offset nonnegative")
	}
	if err == nil && command != "api" && command != "schema" && command != "datasets import" && command != "datasets upload" && !strings.HasPrefix(command, "parse ") && len(rest) > 0 {
		err = cliFail(2, "unexpected arguments for %s", command)
	}
	if err == nil && o.format != "json" && o.format != "jsonl" && o.format != "table" && (command != "datasets export" || o.format != "csv") {
		err = cliFail(2, "unsupported output format %q", o.format)
	}
	ctx, cancel := context.WithTimeout(parent, max(o.timeout, time.Nanosecond))
	defer cancel()
	if err == nil {
		level := slog.LevelWarn
		if o.verbose {
			level = slog.LevelDebug
		}
		if o.quiet {
			level = slog.LevelError
		}
		previous := slog.Default()
		slog.SetDefault(slog.New(slog.NewTextHandler(stderr, &slog.HandlerOptions{Level: level})))
		defer slog.SetDefault(previous)
		if _, ok := cliCommands[command]; !ok {
			err = cliFail(2, "unknown command %q; use fmdash help", command)
		} else {
			envelope.Data, err = cliExecute(ctx, command, rest, o, stdin, &envelope)
		}
	}
	if ctx.Err() != nil {
		err = ctx.Err()
	}
	code := 0
	if err != nil {
		code = 1
		var failure *cliError
		if errors.As(err, &failure) {
			code = failure.code
		}
		if errors.Is(err, errStrictParse) {
			code = 3
		}
		if errors.Is(err, context.DeadlineExceeded) {
			code = 124
		}
		if errors.Is(err, context.Canceled) {
			code = 130
		}
		envelope.Status = "error"
		envelope.Diagnostics = append(envelope.Diagnostics, ParseDiagnostic{Code: "command_failed", Severity: "error", Message: err.Error()})
	}
	var data []byte
	if o.format == "csv" && code == 0 {
		data, _ = envelope.Data.([]byte)
	} else if o.format == "jsonl" {
		data, err = cliJSONLines(envelope)
	} else if o.format == "table" {
		data, err = cliTable(envelope)
	} else {
		data, err = json.MarshalIndent(envelope, "", "  ")
	}
	if err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}
	data = append(data, '\n')
	if o.output != "" && (code == 0 || command == "parse validate" || command == "parse bench" || command == "parse diff") {
		err = cliWriteFile(o.output, data, o.overwrite)
	} else {
		_, err = stdout.Write(data)
	}
	if err != nil {
		_, _ = fmt.Fprintln(stderr, "output:", err)
		return 6
	}
	return code
}

func cliExecute(ctx context.Context, command string, args []string, o cliOptions, stdin io.Reader, e *cliEnvelope) (any, error) {
	switch command {
	case "help", "capabilities":
		return map[string]any{"commands": cliCommands, "usage": "fmdash <command> [subcommand] [input] [flags]", "flags": "fmdash schema", "nativeSaveCapability": "strings only; not player decoding"}, nil
	case "version":
		return cliBuildInfo(), nil
	case "schema":
		var flags []map[string]string
		cliFlagSet(&cliOptions{}).VisitAll(func(f *flag.Flag) {
			flags = append(flags, map[string]string{"name": f.Name, "default": f.DefValue, "description": f.Usage})
		})
		return map[string]any{"schemaVersion": 1, "flags": flags, "envelope": []string{"schemaVersion", "command", "status", "data", "diagnostics", "provenance"}, "exitCodes": map[string]int{"success": 0, "internal": 1, "usage": 2, "invalidInput": 3, "gate": 4, "unavailable": 5, "io": 6, "conflict": 7, "deadline": 124, "interrupted": 130}}, nil
	case "api":
		return cliAPI(ctx, args, o, stdin)
	case "datasets upload":
		if len(args) != 1 {
			return nil, cliFail(2, "datasets upload requires one export path")
		}
		o.upload = args[0]
		o.method = "POST"
		return cliAPI(ctx, []string{"/api/upload"}, o, stdin)
	}
	if path, method, ok := cliRemoteOperation(command, o); ok {
		if command == "cache clear" {
			if err := cliRequired(o.name, "--name cache path"); err != nil {
				return nil, err
			}
		}
		if strings.Contains(path, "{dataset}") {
			if err := cliRequired(o.dataset, "--dataset (backend Dataset identifier)"); err != nil {
				return nil, err
			}
			path = strings.ReplaceAll(path, "{dataset}", o.dataset)
		}
		if o.method == "GET" && method != "" {
			o.method = method
		}
		return cliAPI(ctx, []string{path}, o, stdin)
	}
	if err := cliInitializeConfig(o); err != nil {
		return nil, err
	}
	configData, _ := json.Marshal(map[string]any{"attributeWeights": GetAttributeWeightsSnapshot(), "roleWeights": roleSpecificOverallWeights, "scaled": GetUseScaledRatings()})
	e.Provenance["configurationSha256"] = cliHash(configData)
	if strings.HasPrefix(command, "parse ") {
		return cliParse(ctx, command, args, o, stdin, e)
	}
	return cliWorkflow(ctx, command, args, o, stdin, e)
}
