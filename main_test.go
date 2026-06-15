package main

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"groupie-tracker/api"
)

func init() {
	// Inject full test mocks to protect execution bounds
	registry = &api.UnifiedRegistry{
		Artists: []api.Artist{
			{ID: 1, Name: "Queen", Members: []string{"Freddie Mercury", "Brian May"}, CreationDate: 1970, FirstAlbum: "13-07-1973"},
			{ID: 2, Name: "Gorillaz", FirstAlbum: "26-03-2001"},
		},
		Relations: map[int]api.Relation{
			1: {ID: 1, DatesLocations: map[string][]string{"london-uk": {"20-08-2019"}}},
		},
	}
	templates = template.Must(template.New("index.html").Parse(`{{range .}}<div>{{.Name}}</div>{{end}}`))
	_, _ = templates.New("details.html").Parse(`<h1>{{.Artist.Name}}</h1>`)
}

func TestRouterScenarios(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
	}{
		{"Valid Home GET Request", "GET", "/", http.StatusOK},
		{"Invalid URL Path 404 Protected", "GET", "/invalid-route-path", http.StatusNotFound},
		{"Valid Details Parameter Path", "GET", "/artist?id=1", http.StatusOK},
		{"Missing Artist Profile 404 Matching", "GET", "/artist?id=999", http.StatusNotFound},
		{"Malformed Request Parameters 400 Matching", "GET", "/artist?id=badparam", http.StatusBadRequest},
		{"Valid Search REST Hook Input", "GET", "/api/search?q=Queen", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, nil)
			rr := httptest.NewRecorder()

			if tt.url == "/" || strings.HasPrefix(tt.url, "/invalid") {
				homeHandler(rr, req)
			} else if strings.HasPrefix(tt.url, "/artist") {
				artistDetailsHandler(rr, req)
			} else if strings.HasPrefix(tt.url, "/api/search") {
				apiSearchHandler(rr, req)
			}

			if rr.Code != tt.expectedStatus {
				t.Errorf("%s failed: expected HTTP code %d, received %d", tt.name, tt.expectedStatus, rr.Code)
			}
		})
	}
}