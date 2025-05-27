// src/api/main.go
package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	_ "net/http/pprof" // For profiling, if needed
	"os"
	"path/filepath"
	"strings"
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


func main() {
	// Initialize OpenTelemetry (tracing and metrics) if enabled
	var cleanup func(context.Context) error
	if otelEnabled {
		cleanup = initOTel()
		defer cleanup(context.Background())
		slog.Info("OpenTelemetry initialized", "logs_streaming", true)
		
		// Demo structured logging that will be streamed to OTLP
		DemoStructuredLogging()
	} else {
		slog.Info("OpenTelemetry disabled", "logs_streaming", false)
	}

	// Initialize storage system
	InitStore()
	
	// Start automatic cleanup scheduler for old datasets
	StartCleanupScheduler()
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

	// API endpoint for file uploads
	http.Handle("/upload", wrapHandler(http.HandlerFunc(uploadHandler), "upload"))

	// API endpoint for retrieving player data
	http.Handle("/api/players/", wrapHandler(http.HandlerFunc(playerDataHandler), "player-data"))

	// New API endpoint for retrieving available roles
	http.Handle("/api/roles", wrapHandler(http.HandlerFunc(rolesHandler), "roles"))

	// API endpoint for retrieving leagues data
	http.Handle("/api/leagues/", wrapHandler(http.HandlerFunc(leaguesHandler), "leagues"))

	// API endpoint for retrieving teams data for a specific league
	http.Handle("/api/teams/", wrapHandler(http.HandlerFunc(teamsHandler), "teams"))

	// API endpoint for updating player percentiles with division filtering
	http.Handle("/api/percentiles/", wrapHandler(http.HandlerFunc(percentilesHandler), "percentiles"))

	// Determine port for HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8091" // Default port for the Go API
	}

	slog.Info("Server starting", "port", port, "url", "http://localhost:"+port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe Error: ", err)
	}
}
