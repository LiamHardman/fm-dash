package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "net/http/pprof" // For profiling, if needed
)

func main() {
	// Serve the main index.html page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		// Ensure index.html is in the correct path relative to the executable,
		// or use an absolute path if necessary.
		// Assuming index.html is at the same level as the 'public' directory,
		// and the executable runs from the module root.
		http.ServeFile(w, r, filepath.Join(".", "public", "index.html"))
	})

	// Serve static files from the "public" directory (e.g., CSS, JS, weight JSONs if accessed directly)
	// The JSON weight files are loaded by the backend at startup, not typically served directly unless intended.
	fsPublic := http.FileServer(http.Dir("./public"))
	http.Handle("/public/", http.StripPrefix("/public/", fsPublic))

	// API endpoint for file uploads
	http.HandleFunc("/upload", uploadHandler) // Defined in handlers.go

	// API endpoint for retrieving player data
	http.HandleFunc("/api/players/", playerDataHandler) // Defined in handlers.go

	// Determine port for HTTP server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8091" // Default port
	}

	log.Printf("Server starting on http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe Error: ", err)
	}
}
