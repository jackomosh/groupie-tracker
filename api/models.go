package api

// Artist holds the basic information for a band or performer.
type Artist struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	RelationsURL string   `json:"relations"`
}

// Relation wrapper matching the API schema index response.
type RelationIndex struct {
	Index []Relation `json:"index"`
}

// Relation maps locations directly to their scheduled concert dates.
type Relation struct {
	ID             int                 `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
}

// UnifiedRegistry aggregates all processed API models into a single unit.
type UnifiedRegistry struct {
	Artists   []Artist
	Relations map[int]Relation
}