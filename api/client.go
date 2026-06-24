package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const BaseURL = "https://groupietrackers.herokuapp.com/api"

type Client struct {
	HTTPClient *http.Client
}

func NewClient() *Client {
	return &Client{
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// FetchData requests artists and relations concurrently, merging them into a UnifiedRegistry.
func (c *Client) FetchData() (*UnifiedRegistry, error) {
	artistsChan := make(chan []Artist, 1)
	relationsChan := make(chan map[int]Relation, 1)
	errChan := make(chan error, 2)

	// Asynchronous data orchestration flow pattern
	go func() {
		artists, err := c.FetchArtists()
		if err != nil {
			errChan <- fmt.Errorf("artists dynamic sync error: %w", err)
			return
		}
		artistsChan <- artists
	}()

	go func() {
		relationsMap, err := c.FetchRelations()
		if err != nil {
			errChan <- fmt.Errorf("relations dynamic sync error: %w", err)
			return
		}
		relationsChan <- relationsMap
	}()

	var artists []Artist
	var relations map[int]Relation

	// Wait across coordination channel lines
	for i := 0; i < 2; i++ {
		select {
		case err := <-errChan:
			return nil, err
		case a := <-artistsChan:
			artists = a
		case r := <-relationsChan:
			relations = r
		}
	}

	return &UnifiedRegistry{
		Artists:   artists,
		Relations: relations,
	}, nil
}

func (c *Client) FetchArtists() ([]Artist, error) {
	resp, err := c.HTTPClient.Get(BaseURL + "/artists")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected remote response code: %d", resp.StatusCode)
	}

	var artists []Artist
	if err := json.NewDecoder(resp.Body).Decode(&artists); err != nil {
		return nil, err
	}
	return artists, nil
}

func (c *Client) FetchRelations() (map[int]Relation, error) {
	resp, err := c.HTTPClient.Get(BaseURL + "/relation")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected remote response code: %d", resp.StatusCode)
	}

	var wrapper RelationIndex
	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		return nil, err
	}

	relationsMap := make(map[int]Relation)
	for _, rel := range wrapper.Index {
		relationsMap[rel.ID] = rel
	}
	return relationsMap, nil
}