package http

import "time"

type PostDepsRequest struct {
	Name string `json:"name"`
}

type DepsResponse struct {
	ID int64	`json:"id"`
	Name	string `json:"name"`
	Version string	`json:"version"`
	Dependencies []DependencyNode `json:"dependencies"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
}

type DependencyNode struct {
	Name string `json:"name"`
	Version string `json:"version"`
	Relation string `json:"relation"`
	Score 	*float64 `json:"score,omitempty"`
}

