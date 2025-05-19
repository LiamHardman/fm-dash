package main

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath" // Added for serving files
	"strings"

	"golang.org/x/net/html"
)

// Player struct to hold extracted player data
type Player struct {
	Name       string            `json:"name"`
	Position   string            `json:"position"`
	Age        string            `json:"age"`
	Club       string            `json:"club"`
	Attributes map[string]string `json:"attributes"`
}

// getNodeText extracts text from an HTML node and its children.
// It trims whitespace and joins text from child nodes with single spaces.
func getNodeText(n *html.Node) string {
	if n == nil {
		return ""
	}
	if n.Type == html.TextNode {
		return strings.TrimSpace(n.Data)
	}

	var parts []string
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text := getNodeText(c)
		if text != "" {
			parts = append(parts, text)
		}
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

// uploadHandler handles file uploads, parses HTML, and returns player data as JSON.
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Error parsing multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("playerFile")
	if err != nil {
		http.Error(w, "Error retrieving the file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("Uploaded File: %+v\n", handler.Filename)
	log.Printf("File Size: %+v\n", handler.Size)
	log.Printf("MIME Header: %+v\n", handler.Header)

	doc, err := html.Parse(file)
	if err != nil {
		http.Error(w, "Error parsing HTML: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var players []Player
	var headers []string
	var tableNode *html.Node

	var findTable func(*html.Node)
	findTable = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			tableNode = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findTable(c)
			if tableNode != nil {
				return
			}
		}
	}

	findTable(doc)

	if tableNode == nil {
		http.Error(w, "No table found in the HTML", http.StatusInternalServerError)
		return
	}

	for tr := tableNode.FirstChild; tr != nil; tr = tr.NextSibling {
		if tr.Type == html.ElementNode && tr.Data == "tbody" {
			for rowNode := tr.FirstChild; rowNode != nil; rowNode = rowNode.NextSibling {
				if rowNode.Type == html.ElementNode && rowNode.Data == "tr" {
					processRow(rowNode, &headers, &players)
				}
			}
			break
		} else if tr.Type == html.ElementNode && tr.Data == "tr" {
			processRow(tr, &headers, &players)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*") // For development
	if err := json.NewEncoder(w).Encode(players); err != nil {
		http.Error(w, "Error encoding JSON: "+err.Error(), http.StatusInternalServerError)
	}
}

func processRow(tr *html.Node, headers *[]string, players *[]Player) {
	var cells []string
	isHeaderRow := false

	for td := tr.FirstChild; td != nil; td = td.NextSibling {
		if td.Type == html.ElementNode {
			if td.Data == "th" {
				isHeaderRow = true
				cells = append(cells, getNodeText(td))
			} else if td.Data == "td" {
				cells = append(cells, getNodeText(td))
			}
		}
	}

	if isHeaderRow {
		*headers = cells
		log.Printf("Parsed Headers: %v", *headers)
	} else if len(*headers) > 0 {
		nameIdx := getHeaderIndex(*headers, "Name")
		posIdx := getHeaderIndex(*headers, "Position")
		ageIdx := getHeaderIndex(*headers, "Age")
		clubIdx := getHeaderIndex(*headers, "Club")

		if nameIdx == -1 || nameIdx >= len(cells) || cells[nameIdx] == "" {
			log.Printf("Skipping row, no name found or name index out of bounds. Cells: %v", cells)
			return
		}

		maxRequiredBaseIdx := 0
		for _, idx := range []int{nameIdx, posIdx, ageIdx, clubIdx} {
			if idx > maxRequiredBaseIdx {
				maxRequiredBaseIdx = idx
			}
		}
		if len(cells) <= maxRequiredBaseIdx {
			log.Printf("Skipping row, not enough cells for core data (Name, Position, Age, Club). Cells: %v, Required Max Index: %d", cells, maxRequiredBaseIdx)
			return
		}

		player := Player{
			Name:       safeGet(cells, nameIdx),
			Position:   safeGet(cells, posIdx),
			Age:        safeGet(cells, ageIdx),
			Club:       safeGet(cells, clubIdx),
			Attributes: make(map[string]string),
		}

		attrStartIndex := getHeaderIndex(*headers, "Acc")
		attrEndIndex := getHeaderIndex(*headers, "Pen")

		if attrStartIndex != -1 {
			if attrEndIndex == -1 || attrEndIndex < attrStartIndex {
				log.Printf("'Pen' attribute not found or found before 'Acc'. Trying 'Wor' as fallback end for attributes.")
				fallbackEndIndex := getHeaderIndex(*headers, "Wor")
				if fallbackEndIndex != -1 && fallbackEndIndex >= attrStartIndex {
					attrEndIndex = fallbackEndIndex
				} else {
					log.Printf("'Wor' also not suitable as attribute end. Attributes might be incomplete.")
					attrEndIndex = -1
				}
			}

			if attrEndIndex != -1 {
				for i := attrStartIndex; i <= attrEndIndex && i < len(cells) && i < len(*headers); i++ {
					attrName := (*headers)[i]
					attrValue := cells[i]
					if attrName != "" && attrValue != "" && attrValue != "-" {
						player.Attributes[attrName] = attrValue
					}
				}
			} else {
				log.Printf("Could not determine a valid end for attribute parsing (checked Pen, Wor). Player: %s", player.Name)
			}
		} else {
			log.Printf("'Acc' attribute header not found. Cannot parse detailed attributes for player: %s", player.Name)
		}

		if player.Name != "" {
			*players = append(*players, player)
		} else {
			log.Printf("Skipping row as player name is empty after processing. Cells: %v", cells)
		}

	} else if len(cells) > 1 && cells[1] != "" {
		log.Printf("Processing a row with potentially incomplete data (headers might not be fully parsed yet, or few cells): %v", cells)
	}
}

func getHeaderIndex(headers []string, headerName string) int {
	for i, h := range headers {
		if h == headerName {
			return i
		}
	}
	return -1
}

func safeGet(slice []string, index int) string {
	if index >= 0 && index < len(slice) {
		return slice[index]
	}
	return ""
}

func main() {
	// Serve the index.html file at the root URL "/"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Ensure the request is for the root path, otherwise return 404
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		// Serve the index.html file.
		// This assumes index.html is in the same directory as the executable.
		// For a more robust solution, you might want to use an absolute path
		// or embed the file using Go 1.16+ embed feature.
		http.ServeFile(w, r, filepath.Join(".", "index.html"))
	})

	// Handle for the file upload and parsing
	http.HandleFunc("/upload", uploadHandler)

	log.Println("Server started on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
