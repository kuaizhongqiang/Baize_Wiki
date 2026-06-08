package model

// Config holds all Baize Wiki configuration.
type Config struct {
	Name        string         `yaml:"name" json:"name"`
	Description string         `yaml:"description,omitempty" json:"description,omitempty"`
	Scan        ScanConfig     `yaml:"scan" json:"scan"`
	Output      OutputConfig   `yaml:"output" json:"output"`
	Organize    OrganizeConfig `yaml:"organize" json:"organize"`
	Features    FeatureConfig  `yaml:"features" json:"features"`
	Vector      VectorConfig   `yaml:"vector" json:"vector"`
}

// ScanConfig controls source scanning behaviour.
type ScanConfig struct {
	Paths   []string `yaml:"paths" json:"paths"`
	Exclude []string `yaml:"exclude" json:"exclude"`
	MaxSize int64    `yaml:"max_size" json:"max_size"`
}

// OutputConfig controls Wiki output structure.
type OutputConfig struct {
	Dir   string `yaml:"dir" json:"dir"`
	Level int    `yaml:"level" json:"level"`
	Clean bool   `yaml:"clean" json:"clean"`
}

// OrganizeConfig controls how pages are organized.
type OrganizeConfig struct {
	By string `yaml:"by" json:"by"`
}

// FeatureConfig toggles optional features.
type FeatureConfig struct {
	Draft   bool `yaml:"draft" json:"draft"`
	ScanAll bool `yaml:"scan_all" json:"scan_all"`
	Vector  bool `yaml:"vector" json:"vector"`
}

// VectorConfig controls vector search (Phase 4+).
type VectorConfig struct {
	Mode         string  `yaml:"mode" json:"mode"`                   // local | remote
	HybridWeight float64 `yaml:"hybrid_weight" json:"hybrid_weight"` // BM25 weight α (0.0-1.0)
	Provider     string  `yaml:"provider,omitempty" json:"provider,omitempty"`
	Endpoint     string  `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`     // API endpoint (for remote mode)
	APIKey       string  `yaml:"api_key,omitempty" json:"api_key,omitempty"`       // API key
	Model        string  `yaml:"model,omitempty" json:"model,omitempty"`           // model name for remote API
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Output: OutputConfig{
			Dir:   "./wiki",
			Level: 2,
		},
		Scan: ScanConfig{
			Paths:   []string{"./"},
			MaxSize: 10 * 1024 * 1024, // 10MB
		},
		Organize: OrganizeConfig{
			By: "directory",
		},
		Vector: VectorConfig{
			Mode:         "local",
			HybridWeight: 0.5,
		},
	}
}

// Validate checks config fields and returns an error if invalid.
func (c Config) Validate() error {
	if c.Output.Level < 1 || c.Output.Level > 3 {
		return ErrInvalidConfig
	}
	return nil
}

// Merge applies non-zero flag values on top of the config.
func (c Config) Merge(level int, outputDir string, scanAll bool) Config {
	if level >= 1 && level <= 3 {
		c.Output.Level = level
	}
	if outputDir != "" {
		c.Output.Dir = outputDir
	}
	if scanAll {
		c.Features.ScanAll = true
	}
	return c
}
