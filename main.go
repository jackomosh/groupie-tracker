package main

import (
	"context"
	"errors"
	"fmt"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"groupie-tracker/api"
)

var (
	registry  *api.UnifiedRegistry
	templates *template.Template
)

func main() {
	// Initialize API dependencies cleanly
	client := api.NewClient()
	log.Println("Downloading external tracker metrics into runtime environment memory...")
	data, err := client.FetchData()
	if err != nil {
		log.Fatalf("Fatal system synchronization crash: %v", err)
	}
	registry = data

	// Compile tracking views
	templates = template.Must(template.ParseGlob("templates/*.html"))

	mux := http.NewServeMux()
	
	// Secure routing controls
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/artist", artistDetailsHandler)
	mux.HandleFunc("/api/search", apiSearchHandler)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown strategy
	go func() {
		fmt.Println("Server running smoothly at http://localhost:8080")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Unexpected listener termination: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Stopping runtime operations gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to exit abruptly: %v", err)
	}
	log.Println("Server offline.")
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 Not Found", http.StatusNotFound)
		return
	}
	if err := templates.ExecuteTemplate(w, "index.html", registry.Artists); err != nil {
		http.Error(w, "500 Internal Error", http.StatusInternalServerError)
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

	relation := registry.Relations[id]

	data := struct {
		Artist   api.Artist
		Relation api.Relation
	}{
		Artist:   *targetArtist,
		Relation: relation,
	}

	if err := templates.ExecuteTemplate(w, "details.html", data); err != nil {
		http.Error(w, "500 Internal Error", http.StatusInternalServerError)
	}
}

// Client-Server Event: Real-time API query filter backend engine
func apiSearchHandler(w http.ResponseWriter, r *http.Request) {
	query := strings.ToLower(r.URL.Query().Get("q"))
	w.Header().Set("Content-Type", "application/json")

	// Standard data aggregation algorithm mapping partial inputs
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

	// Fallback to empty array allocation bounds instead of returning null values
	if filtered == nil {
		filtered = []api.Artist{}
	}

	_ = json.NewEncoder(w).Encode(filtered)
}