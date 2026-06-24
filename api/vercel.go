package handler

import (
	"html/template"
	"net/http"
	"strings"
	"sync"

	// Replace "groupie-tracker" with the exact name declared in your go.mod
	"groupie-tracker/api" 
)

var (
	registry *api.UnifiedRegistry
	tmpl     *template.Template
	once     sync.Once
	initErr  error
)

// initializeDependencies ensures data is downloaded exactly once when the serverless container wakes up
func initializeDependencies() {
	client := api.NewClient()
	registry, initErr = client.FetchData()
	if initErr != nil {
		return
	}

	// Compile tracking views using relative directory patterns for Vercel
	tmpl, initErr = template.ParseGlob("../templates/*.html")
}

// Handler is the entry point Vercel targets to process incoming HTTP requests
func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(initializeDependencies)

	if initErr != nil {
		http.Error(w, "500 Internal Synchronisation Error: "+initErr.Error(), http.StatusInternalServerError)
		return
	}

	// Simple Serverless Mux Routing Pattern
	path := r.URL.Path
	if path == "/" {
		homeHandler(w, r)
	} else if path == "/artist" {
		artistDetailsHandler(w, r)
	} else if path == "/api/search" {
		apiSearchHandler(w, r)
	} else {
		http.Error(w, "404 Route Not Found", http.StatusNotFound)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.ExecuteTemplate(w, "index.html", registry.Artists); err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
	}
}

func artistDetailsHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	var targetArtist *api.Artist

	for _, art := range registry.Artists {
		if string(strconv.Itoa(art.ID)) == idStr || art.Name == idStr {
			targetArtist = &art
			break
		}
	}

	if targetArtist == nil {
		http.Error(w, "404 Artist Profile Missing", http.StatusNotFound)
		return
	}

	data := struct {
		Artist   api.Artist
		Relation api.Relation
	}{
		Artist:   *targetArtist,
		Relation: registry.Relations[targetArtist.ID],
	}

	if err := tmpl.ExecuteTemplate(w, "details.html", data); err != nil {
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
	}
}

func apiSearchHandler(w http.ResponseWriter, r *http.Request) {
	query := strings.ToLower(r.URL.Query().Get("q"))
	w.Header().Set("Content-Type", "application/json")

	var filtered []api.Artist
	for _, art := range registry.Artists {
		if strings.Contains(strings.ToLower(art.Name), query) {
			filtered = append(filtered, art)
			continue
		}
		for _, member := range art.Members {
			if strings.Contains(strings.ToLower(member), query) {
				filtered = append(filtered, art)
				break
			}
		}
	}

	if filtered == nil {
		filtered = []api.Artist{}
	}

	_ = json.NewEncoder(w).Encode(filtered)
}