package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
)

// Store handles file system read/write operations for Wiki data.
type Store struct{}

// NewStore creates a new Store.
func NewStore() *Store {
	return &Store{}
}

// WritePage writes a page's content to the file system.
// It ensures parent directories exist and performs atomic writes
// (write to .tmp file, then rename).
func (s *Store) WritePage(wikiDir, pagePath, content string) error {
	fullPath := filepath.Join(wikiDir, pagePath)

	// Ensure parent directory exists
	parentDir := filepath.Dir(fullPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("create directory %s: %w", parentDir, err)
	}

	// Atomic write: write to .tmp then rename
	tmpPath := fullPath + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write temp file %s: %w", tmpPath, err)
	}

	if err := os.Rename(tmpPath, fullPath); err != nil {
		// Clean up temp file on error
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename %s to %s: %w", tmpPath, fullPath, err)
	}

	return nil
}

// WriteMeta writes Wiki metadata to .baize/meta.json.
func (s *Store) WriteMeta(wikiDir string, wiki *model.Wiki) error {
	metaDir := filepath.Join(wikiDir, ".baize")
	if err := os.MkdirAll(metaDir, 0755); err != nil {
		return fmt.Errorf("create .baize directory: %w", err)
	}

	metaPath := filepath.Join(metaDir, "meta.json")
	data, err := json.MarshalIndent(wiki, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal meta: %w", err)
	}

	// Atomic write
	tmpPath := metaPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("write meta temp: %w", err)
	}

	if err := os.Rename(tmpPath, metaPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename meta: %w", err)
	}

	return nil
}

// ReadMeta reads Wiki metadata from .baize/meta.json.
func (s *Store) ReadMeta(wikiDir string) (*model.Wiki, error) {
	metaPath := filepath.Join(wikiDir, ".baize", "meta.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, model.ErrWikiNotFound
		}
		return nil, fmt.Errorf("read meta: %w", err)
	}

	var wiki model.Wiki
	if err := json.Unmarshal(data, &wiki); err != nil {
		return nil, fmt.Errorf("unmarshal meta: %w", err)
	}

	return &wiki, nil
}
