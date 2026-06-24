package main

import (
	"context"
	"embed" // <-- REQUIRED FOR EMBEDDING ASSETS
	"encoding/json"
	"errors"
	"fmt"
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

//go:embed templates/*.html static/**/* static/*
var embeddedFileSystem embed.FS

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

	// Compile tracking views directly out of the embedded binary filesystem context
	templates = template.Must(template.ParseFS(embeddedFileSystem, "templates/*.html"))

	mux := http.NewServeMux()
	
	// Server static assets safely straight from binary memory
	mux.Handle("/static/", http.FileServer(http.FS(embeddedFileSystem)))
	
	mux.HandleFunc("/", homeHandler)
	mux.HandleFunc("/artist", artistDetailsHandler)
	mux.HandleFunc("/api/search", apiSearchHandler)

	// Vercel routes traffic using $PORT environment allocations dynamically
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown strategy
	go func() {
		fmt.Printf("Server running smoothly on port %s\n", port)
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
	// Target file name without directory prefixes because it maps cleanly into the FS
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