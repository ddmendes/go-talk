package model

import (
	"gorm.io/gorm"
)

// Repository model
type Repository struct {
	gorm.Model
	Owner       string
	Name        string
	Stars       int
	Forks       int
	Subscribers int
	Watchers    int
	OpenIssues  int
	Size        int
}
