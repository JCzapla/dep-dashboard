package domain

import "time"

type PackageRef struct {
	Name string `json:"name"`
	Version string `json:"version"`
}

type DependencyNode struct {
	Name string
	Version string
	Relation string
	Score *float64
}

type Package struct {
	ID int64
	PackageRef PackageRef
	Dependencies []DependencyNode
	LastUpdatedAt time.Time
}