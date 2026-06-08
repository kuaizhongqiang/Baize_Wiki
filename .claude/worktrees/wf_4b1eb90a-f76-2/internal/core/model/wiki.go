package model

import "time"

// Wiki represents a complete Wiki knowledge base.
type Wiki struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	SourcePath  string    `json:"source_path"`
	OutputPath  string    `json:"output_path"`
	Config      Config    `json:"config"`
	PageCount   int       `json:"page_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Version     int       `json:"version"`
}

// NewWiki creates a new Wiki with default values.
func NewWiki(name, sourcePath, outputPath string, cfg Config) *Wiki {
	return &Wiki{
		Name:       name,
		SourcePath: sourcePath,
		OutputPath: outputPath,
		Config:     cfg,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Version:    1,
	}
}
