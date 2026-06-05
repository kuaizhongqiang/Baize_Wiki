package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigMergeLevelZero(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, 2, cfg.Output.Level)

	// Merge with level=0 should NOT overwrite the default
	merged := cfg.Merge(0, "")
	assert.Equal(t, 2, merged.Output.Level, "level=0 should not overwrite existing value")

	// Merge with level=1 should overwrite
	merged = cfg.Merge(1, "")
	assert.Equal(t, 1, merged.Output.Level)
}

func TestConfigValidateEdge(t *testing.T) {
	// Valid levels
	for _, l := range []int{1, 2, 3} {
		cfg := Config{Output: OutputConfig{Level: l}}
		assert.NoError(t, cfg.Validate(), "level %d should be valid", l)
	}

	// Invalid levels
	for _, l := range []int{0, 4, -1, 100} {
		cfg := Config{Output: OutputConfig{Level: l}}
		assert.ErrorIs(t, cfg.Validate(), ErrInvalidConfig, "level %d should be invalid", l)
	}
}

func TestPageIDConsistency(t *testing.T) {
	// Same path should always produce the same ID
	id1 := PageID("guide/start.md")
	id2 := PageID("guide/start.md")
	assert.Equal(t, id1, id2, "same path → same ID")

	// Different paths should produce different IDs
	id3 := PageID("guide/end.md")
	assert.NotEqual(t, id1, id3, "different paths → different IDs")
}

func TestWikiVersion(t *testing.T) {
	wiki := NewWiki("test", "/src", "/out", DefaultConfig())
	assert.Equal(t, 1, wiki.Version, "Version should start at 1")
}
