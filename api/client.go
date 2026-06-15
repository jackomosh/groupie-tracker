package api

import (
	"encoding/json"
	"fmt"
	"os"
)

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

// FetchData aggregates files from the disk cache directly
func (c *Client) FetchData() (*UnifiedRegistry, error) {
	artists, err := c.FetchArtists()
	if err != nil {
		return nil, fmt.Errorf("offline artists load failed: %w", err)
	}

	relationsMap, err := c.FetchRelations()
	if err != nil {
		return nil, fmt.Errorf("offline relations load failed: %w", err)
	}

	return &UnifiedRegistry{
		Artists:   artists,
		Relations: relationsMap,
	}, nil
}

func (c *Client) FetchArtists() ([]Artist, error) {
	fileData, err := os.ReadFile("data/artists.json")
	if err != nil {
		return nil, err
	}

	var artists []Artist
	if err := json.Unmarshal(fileData, &artists); err != nil {
		return nil, err
	}
	return artists, nil
}

func (c *Client) FetchRelations() (map[int]Relation, error) {
	fileData, err := os.ReadFile("data/relations.json")
	if err != nil {
		return nil, err
	}

	var wrapper RelationIndex
	if err := json.Unmarshal(fileData, &wrapper); err != nil {
		return nil, err
	}

	relationsMap := make(map[int]Relation)
	for _, rel := range wrapper.Index {
		relationsMap[rel.ID] = rel
	}
	return relationsMap, nil
}