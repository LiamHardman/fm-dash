package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/minio/minio-go/v7"
)

type ColourBand struct {
	Min   int    `json:"min"`
	Color string `json:"color"`
}

type ColourSet struct {
	ID    string       `json:"id"`
	Name  string       `json:"name"`
	Bands []ColourBand `json:"bands"`
}

// Each scale has its own document so saving FM settings cannot overwrite FIFA settings.
type ColourSetDocument struct {
	Version     int         `json:"version"`
	ActiveSetID string      `json:"activeSetId"`
	Sets        []ColourSet `json:"sets"`
}

var colourHexPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func validateColourSetDocument(document *ColourSetDocument, scale string) error {
	minimum, maximum := 0, 100
	if scale == "fm" {
		minimum, maximum = 1, 20
	}
	if document.Version != 1 || document.Sets == nil || len(document.Sets) > 200 {
		return fmt.Errorf("colour sets must use version 1 and contain at most 200 sets")
	}
	seen := make(map[string]bool)
	for i := range document.Sets {
		set := &document.Sets[i]
		set.Name = strings.TrimSpace(set.Name)
		if strings.TrimSpace(set.ID) == "" || len(set.ID) > 100 || seen[set.ID] || set.Name == "" || len([]rune(set.Name)) > 100 {
			return fmt.Errorf("sets need unique IDs and names of 1–100 characters")
		}
		seen[set.ID] = true
		if len(set.Bands) < 1 || len(set.Bands) > 20 || set.Bands[0].Min != minimum {
			return fmt.Errorf("sets need 1–20 bands, starting at %d", minimum)
		}
		previous := minimum - 1
		for _, band := range set.Bands {
			if band.Min <= previous || band.Min > maximum || !colourHexPattern.MatchString(band.Color) {
				return fmt.Errorf("band starts must increase up to %d, with colours in #RRGGBB format", maximum)
			}
			previous = band.Min
		}
	}
	if document.ActiveSetID != "" && !seen[document.ActiveSetID] {
		return fmt.Errorf("active colour set does not exist")
	}
	return nil
}

func colourSetsHandler(w http.ResponseWriter, r *http.Request) {
	scale := strings.TrimPrefix(r.URL.Path, "/api/colour-sets/")
	if scale != "fm" && scale != "fifa" {
		WriteErrorResponse(w, r, "invalid_scale", "Choose fm or fifa", nil, http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodPut {
		w.Header().Set("Allow", "GET, PUT")
		WriteErrorResponse(w, r, "method_not_allowed", "Method not allowed", nil, http.StatusMethodNotAllowed)
		return
	}
	if configStorage == nil {
		WriteErrorResponse(w, r, "storage_unavailable", "Colour-set storage is not initialised", nil, http.StatusServiceUnavailable)
		return
	}
	// Keep durable settings outside the files scanned by Dataset retention cleanup.
	key := "colour-sets/" + scale + ".json"
	document := ColourSetDocument{Version: 1, Sets: []ColourSet{}}
	if r.Method == http.MethodGet {
		data, err := configStorage.RetrieveConfig(key)
		if err != nil {
			if !os.IsNotExist(err) && minio.ToErrorResponse(err).Code != "NoSuchKey" {
				WriteErrorResponse(w, r, "storage_error", "Could not load colour sets", nil, http.StatusInternalServerError)
				return
			}
		} else if json.Unmarshal(data, &document) != nil || validateColourSetDocument(&document, scale) != nil {
			WriteErrorResponse(w, r, "storage_error", "Stored colour sets are invalid", nil, http.StatusInternalServerError)
			return
		}
	} else {
		document = ColourSetDocument{}
		r.Body = http.MaxBytesReader(w, r.Body, 2*1024*1024)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&document); err != nil {
			WriteErrorResponse(w, r, "invalid_json", "Request must be a valid colour-set document (maximum 2 MB)", nil, http.StatusBadRequest)
			return
		}
		if err := decoder.Decode(new(any)); err != io.EOF {
			WriteErrorResponse(w, r, "invalid_json", "Request must contain one JSON document", nil, http.StatusBadRequest)
			return
		}
		if err := validateColourSetDocument(&document, scale); err != nil {
			WriteErrorResponse(w, r, "invalid_colour_set", err.Error(), nil, http.StatusBadRequest)
			return
		}
		data, err := json.Marshal(document)
		if err != nil {
			WriteErrorResponse(w, r, "serialization_error", "Could not encode colour sets", nil, http.StatusInternalServerError)
			return
		}
		if err := configStorage.StoreConfig(key, data); err != nil {
			WriteErrorResponse(w, r, "storage_error", "Could not save colour sets", nil, http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(document); err != nil {
		LogWarn("colour-sets: failed to write response: %v", err)
	}
}
