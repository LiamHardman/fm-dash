// src/api/main.go
package main

import (
	"log"
	"net/http"
	_ "net/http/pprof" // For profiling, if needed
	"os"
	"path/filepath"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Initialize storage system
	InitStore()
	
	// Start automatic cleanup scheduler for old datasets
	StartCleanupScheduler()
	// Serve the main index.html page (assuming it's built into a 'public' or 'dist' folder by Vue)
	// Adjust the path according to your frontend build output.
	// If Vue serves on a different port (e.g., 3000) and proxies API calls, this might not be needed here.
	// However, if Go is serving everything, ensure this path is correct.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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

	// Serve static files from the "public" directory (e.g., CSS, JS from Vue build, weight JSONs)
	// This path also assumes the 'public' folder is relative to where the Go executable is run.
	fsPublic := http.FileServer(http.Dir("./public"))
	http.Handle("/public/", http.StripPrefix("/public/", fsPublic))
	// If your Vue assets are in 'dist/assets', you might need:
	// fsAssets := http.FileServer(http.Dir("./dist/assets"))
	// http.Handle("/assets/", http.StripPrefix("/assets/", fsAssets))

	// API endpoint for file uploads
	http.HandleFunc("/upload", uploadHandler) // Defined in handlers.go

	// API endpoint for retrieving player data
	http.HandleFunc("/api/players/", playerDataHandler) // Defined in handlers.go

	// New API endpoint for retrieving available roles
	http.HandleFunc("/api/roles", rolesHandler) // Defined in handlers.go

	// Prometheus metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	// Determine port for HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8091" // Default port for the Go API
	}

	log.Printf("Server starting on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe Error: ", err)
	}
}
