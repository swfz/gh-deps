package models

import "strings"

// Repository represents a GitHub repository
type Repository struct {
	NameWithOwner string // Full repository name (owner/repo)
	Name          string // Repository name only
	Owner         string // Owner/organization name
}

// NewRepository creates a Repository from a full name (owner/repo)
func NewRepository(nameWithOwner string) Repository {
	parts := strings.Split(nameWithOwner, "/")
	if len(parts) >= 2 {
		return Repository{
			NameWithOwner: nameWithOwner,
			Owner:         parts[0],
			Name:          parts[1],
		}
	}
	return Repository{
		NameWithOwner: nameWithOwner,
		Name:          nameWithOwner,
	}
}
