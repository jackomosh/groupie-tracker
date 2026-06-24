package handler

import (
	"embed"
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"groupie-tracker/api" // Ensure this matches your go.mod module path precisely
)

// Go Embed pulls assets relative to the current file. Since this is inside 'api/', we step up one directory level.
//go:embed ../templates/*.html ../static/**/* ../static/*
var embeddedFileSystem embed.FS

var (
	registry *api.UnifiedRegistry
	tmpl     *template.Template
	once     sync.Once
	initErr  error
)

// initializeDependencies executes exactly once when the serverless container spins up
func initializeDependencies() {
	client := api.NewClient()
	registry, initErr = client.FetchData()
	if initErr != nil {
		return
	}

	// Parse templates out of the embedded filesystem layer safely
	tmpl, initErr = template.ParseFS(embeddedFileSystem, "../templates/*.html")
}

// Handler is the entry point Vercel targets to process incoming serverless calls
func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(initializeDependencies)

	if initErr != nil {
		http.Error(w, "500 Internal Initialization Error: "+initErr.Error(), http.StatusInternalServerError)
		return
	}

	path := r.URL.Path

	// Route Static assets directly out of the embedded binary filesystem
	if strings.HasPrefix(path, "/static/") {
		// Strip the nested directory prefix so it lines up with the embedded structure bounds cleanly
		fs := http.FileServer(http.FS(embeddedFileSystem))
		// Prepend the step-up reference to point to our layout tree structure mapping
		r.URL.Path = "../" + path
		fs.ServeHTTP(w, r)
		return
	}

	// Serverless HTTP Route Mux Matrix
	switch path {
	case "/":
		homeHandler(w, r)
	case "/artist":
		artistDetailsHandler(w, r)
	case "/api/search":
		apiSearchHandler(w, r)
	default:
		http.Error(w, "404 Not Found", http.StatusNotFound)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if err := tmpl.ExecuteTemplate(w, "index.html", registry.Artists); err != nil {
		http.Error(w, "500 Internal Server Error: "+err.Error(), http.StatusInternalServerError)
	}
}

func artistDetailsHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "400 Bad Request Parameters", http.StatusBadRequest)
		return
	}

	var targetArtist *api.Artist
	for _, art := range registry.Artists {
		if art.ID == id {
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
		Relation: registry.Relations[id],
	}

	if err := tmpl.ExecuteTemplate(w, "details.html", data); err != nil {
		http.Error(w, "500 Internal Error", http.StatusInternalServerError)
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