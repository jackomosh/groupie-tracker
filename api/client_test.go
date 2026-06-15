package api

import (
	"os"
	"testing"
)

func TestOfflineFetchData(t *testing.T) {
	// 1. Create a temporary data directory for testing
	err := os.MkdirAll("data", 0755)
	if err != nil {
		t.Fatalf("Failed to create temporary testing data folder: %v", err)
	}

	// 2. Write minimalist valid mock files so the reader doesn't fail
	mockArtists := []byte(`[{"id":1,"name":"Queen","creationDate":1970}]`)
	mockRelations := []byte(`{"index":[{"id":1,"datesLocations":{"london-uk":["20-08-2019"]}}]}`)

	_ = os.WriteFile("data/artists.json", mockArtists, 0644)
	_ = os.WriteFile("data/relations.json", mockRelations, 0644)

	// 3. Execute the client target check
	client := NewClient()
	registry, err := client.FetchData()

	if err != nil {
		t.Fatalf("FetchData failed during testing run: %v", err)
	}

	if len(registry.Artists) != 1 || registry.Artists[0].Name != "Queen" {
		t.Errorf("Expected artist 'Queen', got something else or empty data")
	}

	if _, exists := registry.Relations[1]; !exists {
		t.Errorf("Expected relation with ID 1 to be loaded into the memory map")
	}
}