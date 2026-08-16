package main

import (
	"encoding/json"
	"net/http"
)

// fmSaveMaxUploadSize is deliberately separate from getMaxUploadSize() (which defaults to 20MB
// for CSV/HTML player exports) -- native .fm saves are routinely 200MB+.
const fmSaveMaxUploadSize = 500 * 1024 * 1024

// fmSaveImportHandler handles POST requests uploading a Football Manager .fm save file. It
// decompresses the save and scans a window near the start of the archive for whatever
// length-prefixed strings can be found -- see .scratch/fm-save-parsing/map.md for why this is
// the whole technique: the format is undocumented and self-describing-but-proprietary, so there
// is no known record schema to decode against.
func fmSaveImportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, fmSaveMaxUploadSize)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Error parsing multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("saveFile")
	if err != nil {
		http.Error(w, "Error retrieving the file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	info, err := ParseFMSave(file)
	if err != nil {
		http.Error(w, "Could not parse .fm save: "+err.Error(), http.StatusUnprocessableEntity)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(info); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}
