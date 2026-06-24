package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLiveFetchLogic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"id":1,"name":"Queen","creationDate":1970}]`))
	}))
	defer server.Close()

	client := &Client{
		HTTPClient: &http.Client{Timeout: 2 * time.Second},
	}

	resp, err := client.HTTPClient.Get(server.URL)
	if err != nil {
		t.Fatalf("Mock test call dropped: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.StatusCode)
	}
}