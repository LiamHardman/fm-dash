package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

var (
	otelEnabled = strings.ToLower(getEnvWithDefault("OTEL_ENABLED", "false")) == "true"
)

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// configureRuntimeMemory optimizes Go runtime settings for better memory management
func configureRuntimeMemory() {
	// Statically set GOGC to 150 for consistent performance
	debug.SetGCPercent(150)
	slog.Debug("Statically set GOGC to 150")

	// Set memory limit if specified
	if memLimit := os.Getenv("GOMEMLIMIT"); memLimit != "" {
		debug.SetMemoryLimit(parseMemoryLimit(memLimit))
		slog.Debug("Set memory limit", "limit", memLimit)
	}

	// Configure maximum number of OS threads
	runtime.GOMAXPROCS(runtime.NumCPU())

	slog.Debug("Runtime configured",
		"GOMAXPROCS", runtime.GOMAXPROCS(0),
		"NumCPU", runtime.NumCPU())
}

// parseMemoryLimit parses memory limit strings like "1GB", "512MB"
func parseMemoryLimit(limit string) int64 {
	if limit == "" {
		return -1
	}

	// Simple parser for common formats
	var multiplier int64 = 1

	if len(limit) >= 2 {
		suffix := limit[len(limit)-2:]
		switch suffix {
		case "KB", "kb":
			multiplier = 1024
			limit = limit[:len(limit)-2]
		case "MB", "mb":
			multiplier = 1024 * 1024
			limit = limit[:len(limit)-2]
		case "GB", "gb":
			multiplier = 1024 * 1024 * 1024
			limit = limit[:len(limit)-2]
		}
	}

	if val, err := strconv.ParseInt(limit, 10, 64); err == nil {
		return val * multiplier
	}

	return -1 // Invalid format
}

func validateEnvironmentVariables() error {
	// Validate OTEL_EXPORTER_OTLP_ENDPOINT if set
	if endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); endpoint != "" {
		if !strings.Contains(endpoint, ":") {
			return WrapErrorf(ErrInvalidOTELEndpoint, "endpoint: %s (must include port)", endpoint)
		}
	}

	// Validate S3_ENDPOINT format if set
	if endpoint := os.Getenv("S3_ENDPOINT"); endpoint != "" {
		if !strings.Contains(endpoint, ":") && !strings.HasPrefix(endpoint, "http") {
			return WrapErrorf(ErrInvalidS3Endpoint, "endpoint: %s (should include port or be full URL): invalid S3_ENDPOINT format", endpoint)
		}
	}

	// Validate SERVICE_NAME doesn't contain dangerous characters
	if serviceName := os.Getenv("SERVICE_NAME"); serviceName != "" {
		if strings.ContainsAny(serviceName, " \t\n\r;|&$`") {
			return ErrInvalidServiceName
		}
	}

	return nil
}

func main() {
	// Validate environment variables first
	if err := validateEnvironmentVariables(); err != nil {
		slog.Error("Environment validation failed", "error", err)
		os.Exit(1)
	}

	// Configure Go runtime for better memory management
	configureRuntimeMemory()

	// Determine port for HTTP server with validation
	port := os.Getenv("PORT")
	if port == "" {
		port = "8091" // Default port for the Go API
	} else {
		// Validate port is a valid number and in reasonable range
		if portNum, err := strconv.Atoi(port); err != nil || portNum <= 0 || portNum > 65535 {
			slog.Error("Invalid PORT environment variable. Must be a number between 1-65535", "port", port)
			os.Exit(1)
		}
	}

	// Initialize OpenTelemetry (tracing and metrics) if enabled
	var cleanup func(context.Context) error
	if otelEnabled {
		cleanup = initOTel()
		if cleanup == nil {
			slog.Error("Failed to initialize OpenTelemetry: initOTel returned nil cleanup function")
			os.Exit(1)
		}
		defer func() {
			if err := cleanup(context.Background()); err != nil {
				LogWarn("Warning: Error during OpenTelemetry cleanup: %v", err)
			}
		}()
		slog.Debug("OpenTelemetry initialized", "logs_streaming", true)

	} else {
		slog.Info("OpenTelemetry disabled", "logs_streaming", false)
	}

	// Initialize storage system
	InitStore()

	// Initialize in-memory cache
	InitInMemoryCache()

	// Initialize cache storage system
	InitCacheStorage(context.Background())

	// Initialize memory optimizations
	InitializeMemoryOptimizations()

	// Initialize advanced memory optimizer (starts periodic profiler snapshots
	// consumed by the /api/memory-profile debug endpoints)
	InitializeMemoryProfiler()

	// Start automatic cleanup scheduler for old datasets
	StartCleanupScheduler()

	// Start cache cleanup scheduler for performance optimization
	StartCacheCleanupScheduler()

	// Start async configuration loading in background
	go func() {
		configInitOnce.Do(initializeConfigAsync)
	}()

	// Initialize regen fallback face cache
	initRegenCache()

	// Start performance monitoring (log metrics every 30 seconds)
	StartPerformanceMonitoring(30 * time.Second)
	// Serve the main index.html page (assuming it's built into a 'public' or 'dist' folder by Vue)
	// Adjust the path according to your frontend build output.
	// If Vue serves on a different port (e.g., 3000) and proxies API calls, this might not be needed here.
	// However, if Go is serving everything, ensure this path is correct.
	indexHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			// Fallback to serving static files if not root, or let fsPublic handle it
			// For a Single Page Application, usually all non-API routes serve index.html
			// http.ServeFile(w, r, filepath.Join(".", "public", "index.html")) // Example
			// For now, let's keep it simple: only "/" serves index.html directly.
			// Other static assets are handled by fsPublic.
			// If you have client-side routing, you'll need a more sophisticated setup here.
			http.NotFound(w, r)
			return
		}
		// This path assumes your Go executable is run from the root of the Go API module,
		// and 'public/index.html' is relative to that.
		// If your Vue build output is elsewhere, adjust this path.
		// For example, if Vue builds to '../dist', it might be filepath.Join("..", "dist", "index.html")
		http.ServeFile(w, r, filepath.Join(".", "public", "index.html"))
	})
	http.Handle("/", wrapHandler(indexHandler, "index"))

	// Serve static files from the "public" directory (e.g., CSS, JS from Vue build, weight JSONs)
	// This path also assumes the 'public' folder is relative to where the Go executable is run.
	fsPublic := http.FileServer(http.Dir("./public"))
	http.Handle("/public/", http.StripPrefix("/public/", fsPublic))
	// If your Vue assets are in 'dist/assets', you might need:
	// fsAssets := http.FileServer(http.Dir("./dist/assets"))
	// http.Handle("/assets/", http.StripPrefix("/assets/", fsAssets))

	// Create HTTP server with timeouts and middleware
	mux := http.NewServeMux()

	// 1. Panic Recovery (outermost - catches all panics)
	// 2. Request Context (adds request ID and other context)
	// 3. CORS (early for preflight requests)
	// 4. Compression (before logging to compress logs too)
	// 5. Logging (captures all request/response data)
	// 6. Timeout (inner control)
	var handler http.Handler = mux
	handler = RequestTimeoutMiddlewareFunc(func(r *http.Request) time.Duration {
		// Who to Sign drives a multi-round LLM tool-calling loop against OpenAI, which
		// legitimately takes longer than the app-wide default.
		if strings.HasPrefix(r.URL.Path, "/api/who-to-sign/") {
			return 120 * time.Second
		}
		// Scout Report drives the same shape of multi-round LLM tool-calling loop.
		if strings.HasPrefix(r.URL.Path, "/api/scout-report/") {
			return 120 * time.Second
		}
		// Chatbot drives the same shape of loop per turn, streamed via SSE.
		if strings.HasPrefix(r.URL.Path, "/api/chatbot/") {
			return 120 * time.Second
		}
		return 30 * time.Second
	})(handler) // request timeout, per-route
	handler = LoggingMiddleware(handler)
	handler = CompressionMiddleware(handler) // Add compression middleware
	handler = CORSMiddleware(handler)
	handler = RequestContextMiddleware(handler)
	handler = PanicRecoveryMiddleware(handler)

	// Re-register all routes with the new mux
	mux.Handle("/", wrapHandler(indexHandler, "index"))
	mux.Handle("/public/", http.StripPrefix("/public/", fsPublic))
	mux.Handle("/api/upload", wrapHandler(http.HandlerFunc(uploadHandler), "upload"))
	mux.Handle("/api/upload/preflight", wrapHandler(http.HandlerFunc(uploadPreflightHandler), "upload-preflight"))
	mux.Handle("/api/players/", wrapHandler(GetFormatAwareCacheHandler(), "player-data"))
	mux.Handle("/api/roles", wrapHandler(http.HandlerFunc(cachedRolesHandler), "roles"))
	mux.Handle("/api/leagues/", wrapHandler(http.HandlerFunc(leaguesHandler), "leagues"))
	mux.Handle("/api/divisions/", wrapHandler(http.HandlerFunc(divisionsHandler), "divisions"))
	mux.Handle("/api/teams/", wrapHandler(http.HandlerFunc(teamsHandler), "teams"))
	mux.Handle("/api/top-teams/", wrapHandler(http.HandlerFunc(topTeamsHandler), "top-teams"))
	mux.Handle("/api/nations-summary/", wrapHandler(http.HandlerFunc(nationsSummaryHandler), "nations-summary"))
	mux.Handle("/api/percentiles/", wrapHandler(http.HandlerFunc(percentilesHandler), "percentiles"))
	mux.Handle("/api/player-percentiles/", wrapHandler(http.HandlerFunc(playerPercentilesHandler), "player-percentiles"))
	mux.Handle("/api/percentiles-status/", wrapHandler(http.HandlerFunc(percentilesStatusHandler), "percentiles-status"))
	mux.Handle("/api/search/", wrapHandler(http.HandlerFunc(searchHandler), "search"))
	mux.Handle("/api/config", wrapHandler(http.HandlerFunc(cachedConfigHandler), "config"))
	mux.Handle("/api/bargain-hunter/", wrapHandler(http.HandlerFunc(bargainHunterHandler), "bargain-hunter"))
	mux.Handle("/api/who-to-sign/", wrapHandler(http.HandlerFunc(whoToSignHandler), "who-to-sign"))
	mux.Handle("/api/managed-team/", wrapHandler(http.HandlerFunc(managedTeamHandler), "managed-team"))
	mux.Handle("/api/scout-report-stars/", wrapHandler(http.HandlerFunc(scoutReportStarsHandler), "scout-report-stars"))
	mux.Handle("/api/scout-report/", wrapHandler(http.HandlerFunc(scoutReportHandler), "scout-report"))
	mux.Handle("/api/scout-reports/", wrapHandler(http.HandlerFunc(scoutReportsBookHandler), "scout-reports-book"))
	mux.Handle("/api/chatbot/", wrapHandler(http.HandlerFunc(chatbotHandler), "chatbot"))
	mux.Handle("/api/fm-save-import", wrapHandler(http.HandlerFunc(fmSaveImportHandler), "fm-save-import"))

	// API endpoint for serving player face images
	mux.Handle("/api/faces", wrapHandler(http.HandlerFunc(facesHandler), "faces"))

	// API endpoint for serving team logo images
	mux.Handle("/api/logos", wrapHandler(http.HandlerFunc(logosHandler), "logos"))

	// API endpoint for team name to ID matching
	mux.Handle("/api/team-match", wrapHandler(http.HandlerFunc(teamMatchHandler), "team-match"))

	// API endpoint for resolving a club name to logo display metadata
	mux.Handle("/api/club-logo/resolve", wrapHandler(http.HandlerFunc(clubLogoResolveHandler), "club-logo-resolve"))

	// API endpoint for managing user-confirmed/rejected club logo overrides
	mux.Handle("/api/club-logo/overrides", wrapHandler(http.HandlerFunc(clubLogoOverridesHandler), "club-logo-overrides"))

	// API endpoint for cache operations (nation ratings, etc.)
	mux.Handle("/api/cache/", wrapHandler(http.HandlerFunc(cacheHandler), "cache"))

	// API endpoint for cache status and monitoring
	mux.Handle("/api/cache-status", wrapHandler(http.HandlerFunc(cacheStatusHandler), "cache-status"))

	// API endpoint for memory optimization reports
	mux.Handle("/api/memory-optimization", wrapHandler(http.HandlerFunc(GetMemoryOptimizationHandler()), "memory-optimization"))

	// Debug endpoint for checking JSON serialization
	mux.Handle("/api/debug/player", wrapHandler(http.HandlerFunc(debugPlayerDataHandler), "debug-player"))

	// API endpoint for processing status checks

	// API endpoint for detailed player stats
	mux.Handle("/api/fullplayerstats/", wrapHandler(http.HandlerFunc(fullPlayerStatsHandler), "full-player-stats"))
	mux.Handle("/api/nation-top-players/", wrapHandler(http.HandlerFunc(nationTopPlayersHandler), "nation-top-players"))
	mux.Handle("/api/team-data/", wrapHandler(http.HandlerFunc(teamDataHandler), "team-data"))

	// API endpoint for finding player upgrades
	mux.Handle("/api/upgrade-finder", wrapHandler(http.HandlerFunc(upgradeFinderHandler), "upgrade-finder"))

	// API endpoint for performance data
	mux.Handle("/api/performance/", wrapHandler(http.HandlerFunc(performanceDataHandler), "performance-data"))

	// API endpoint for exporting complete dataset data
	mux.Handle("/api/export/", wrapHandler(http.HandlerFunc(exportDataHandler), "export-data"))

	// API endpoint for persisting per-dataset wishlists
	mux.Handle("/api/wishlists/", wrapHandler(http.HandlerFunc(wishlistHandler), "wishlists"))
	// API endpoint for installation-level, cross-dataset shortlists
	mux.Handle("/api/shortlists", wrapHandler(http.HandlerFunc(shortlistsHandler), "shortlists"))
	// API endpoint for installation-level saved filter recipes
	mux.Handle("/api/saved-searches", wrapHandler(http.HandlerFunc(savedSearchesHandler), "saved-searches"))

	// API endpoint for multi-snapshot player progression analysis
	mux.Handle("/api/progression/analyze", wrapHandler(http.HandlerFunc(progressionAnalyzeHandler), "progression-analyze"))
	mux.Handle("/api/progression/projects", wrapHandler(http.HandlerFunc(progressionProjectsHandler), "progression-projects"))

	// Register memory profiling and pprof endpoints
	RegisterMemoryProfileEndpoints(mux)

	// Create server with proper timeouts.
	//
	// WriteTimeout is a raw connection-level deadline enforced by net/http itself — it
	// starts when request headers finish being read and covers the entire handler
	// execution plus response write, independent of any per-request context.Context
	// deadline set by RequestTimeoutMiddlewareFunc above. It must therefore be at least
	// as large as the longest per-route budget that middleware hands out (currently 120s
	// for /api/who-to-sign/), or net/http will kill the connection out from under a
	// still-running handler well before its context deadline fires — which surfaced as
	// intermittent 502s on who-to-sign requests that legitimately ran past 30s.
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadTimeout:       15 * time.Second,  // Time to read request
		WriteTimeout:      130 * time.Second, // Time to write response — must exceed the largest per-route timeout above
		IdleTimeout:       60 * time.Second,  // Time to keep connection open
		ReadHeaderTimeout: 5 * time.Second,   // Time to read request headers
	}

	slog.Debug("Server starting",
		"port", port,
		"url", "http://localhost:"+port,
		"read_timeout", "15s",
		"write_timeout", "130s",
		"idle_timeout", "60s")

	if err := server.ListenAndServe(); err != nil {
		slog.Error("Server failed to start", "error", err)
		//nolint:gocritic // exitAfterDefer: this is at the end of main, defers have already run
		os.Exit(1)
	}
}
