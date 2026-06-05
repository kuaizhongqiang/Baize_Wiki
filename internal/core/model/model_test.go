package model

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, 2, cfg.Output.Level)
	assert.Equal(t, int64(10*1024*1024), cfg.Scan.MaxSize)
	assert.Equal(t, "./wiki", cfg.Output.Dir)
	assert.Equal(t, "directory", cfg.Organize.By)
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name  string
		level int
		valid bool
	}{
		{"level 1", 1, true},
		{"level 2", 2, true},
		{"level 3", 3, true},
		{"level 0 invalid", 0, false},
		{"level 4 invalid", 4, false},
		{"level -1 invalid", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Output.Level = tt.level
			err := cfg.Validate()
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, ErrInvalidConfig)
			}
		})
	}
}

func TestConfigMerge(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Output.Level = 1

	// Merge with valid level override
	merged := cfg.Merge(3, "")
	assert.Equal(t, 3, merged.Output.Level)
	assert.Equal(t, "./wiki", merged.Output.Dir)

	// Merge with output dir override
	merged2 := cfg.Merge(0, "./custom")
	assert.Equal(t, 1, merged2.Output.Level)
	assert.Equal(t, "./custom", merged2.Output.Dir)

	// Merge with both overrides
	merged3 := cfg.Merge(2, "./other")
	assert.Equal(t, 2, merged3.Output.Level)
	assert.Equal(t, "./other", merged3.Output.Dir)

	// Invalid levels should not be applied by Merge (Validate catches them)
	merged4 := cfg.Merge(0, "")
	assert.Equal(t, 1, merged4.Output.Level)
}

func TestPageID(t *testing.T) {
	id1 := PageID("docs/guide/getting-started.md")
	id2 := PageID("docs/guide/getting-started.md")
	id3 := PageID("docs/guide/advanced-usage.md")

	// Same path produces same ID
	assert.Equal(t, id1, id2)

	// Different path produces different ID
	assert.NotEqual(t, id1, id3)

	// ID is non-empty and starts with page_ prefix
	assert.Contains(t, id1, "page_")
}

func TestNewWiki(t *testing.T) {
	cfg := DefaultConfig()
	w := NewWiki("Test Wiki", "./src", "./out", cfg)

	assert.Equal(t, "Test Wiki", w.Name)
	assert.Equal(t, "./src", w.SourcePath)
	assert.Equal(t, "./out", w.OutputPath)
	assert.Equal(t, 1, w.Version)
	assert.NotZero(t, w.CreatedAt)
	assert.NotZero(t, w.UpdatedAt)
}

func TestSentinelErrors(t *testing.T) {
	assert.True(t, errors.Is(ErrWikiNotFound, ErrWikiNotFound))
	assert.True(t, errors.Is(ErrPageNotFound, ErrPageNotFound))
	assert.True(t, errors.Is(ErrSourceNotFound, ErrSourceNotFound))
	assert.True(t, errors.Is(ErrInvalidConfig, ErrInvalidConfig))
	assert.True(t, errors.Is(ErrScanFailed, ErrScanFailed))
	assert.True(t, errors.Is(ErrGenerateFailed, ErrGenerateFailed))
	assert.True(t, errors.Is(ErrEmptySource, ErrEmptySource))
}

func TestStructuredError(t *testing.T) {
	e := &Error{
		Code:    "ERR_TEST",
		Message: "test error",
		Detail:  "extra info",
	}
	assert.Contains(t, e.Error(), "ERR_TEST")
	assert.Contains(t, e.Error(), "test error")
	assert.Contains(t, e.Error(), "extra info")

	// Without detail
	e2 := &Error{
		Code:    "ERR_SIMPLE",
		Message: "simple error",
	}
	assert.Contains(t, e2.Error(), "ERR_SIMPLE")
}

func TestFrontmatterCustomFields(t *testing.T) {
	fm := Frontmatter{
		Title: "Test",
		Tags:  []string{"go", "test"},
		Custom: map[string]any{
			"version": 2,
			"status":  "active",
		},
	}
	assert.Equal(t, "Test", fm.Title)
	assert.Equal(t, []string{"go", "test"}, fm.Tags)
	assert.Equal(t, 2, fm.Custom["version"])
	assert.Equal(t, "active", fm.Custom["status"])
}
