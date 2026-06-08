package model

// Config holds all Baize Wiki configuration.
type Config struct {
	Name        string         `yaml:"name" json:"name"`
	Description string         `yaml:"description,omitempty" json:"description,omitempty"`
	Profile     string         `yaml:"profile,omitempty" json:"profile,omitempty"` // speed | balanced | local
	Scan        ScanConfig     `yaml:"scan" json:"scan"`
	Output      OutputConfig   `yaml:"output" json:"output"`
	Organize    OrganizeConfig `yaml:"organize" json:"organize"`
	Features    FeatureConfig  `yaml:"features" json:"features"`
	Catalog     CatalogConfig  `yaml:"catalog" json:"catalog"`
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

// CatalogConfig controls the cataloging pipeline (Level 2/3).
type CatalogConfig struct {
	Level    int    `yaml:"level" json:"level"`                         // 0=none, 1=summary, 2=full
	Backend  string `yaml:"backend,omitempty" json:"backend,omitempty"` // local | remote
	Endpoint string `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	Model    string `yaml:"model,omitempty" json:"model,omitempty"`
	APIKey   string `yaml:"api_key,omitempty" json:"api_key,omitempty"`
}

// VectorConfig controls vector search (Phase 4+).
type VectorConfig struct {
	Profile      string  `yaml:"profile,omitempty" json:"profile,omitempty"` // speed | balanced | local
	Mode         string  `yaml:"mode" json:"mode"`                   // local | remote
	HybridWeight float64 `yaml:"hybrid_weight" json:"hybrid_weight"` // BM25 weight α (0.0-1.0)
	Provider     string  `yaml:"provider,omitempty" json:"provider,omitempty"`
	Endpoint     string  `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`     // API endpoint (for remote mode)
	APIKey       string  `yaml:"api_key,omitempty" json:"api_key,omitempty"`       // API key
	Model        string  `yaml:"model,omitempty" json:"model,omitempty"`           // model name for remote API
}

// DefaultConfig returns a Config with sensible defaults (local profile).
func DefaultConfig() Config {
	cfg := Config{
		Profile: "local",
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
		Catalog: CatalogConfig{
			Backend:  "local",
			Endpoint: "http://localhost:1234/v1",
			Model:    "qwen/qwen3.5-9b",
		},
		Vector: VectorConfig{
			Mode:         "local",
			HybridWeight: 0.5,
			Endpoint:     "http://localhost:1234/v1",
			Model:        "text-embedding-baai-bge-m3-568m:2",
		},
	}
	cfg.ApplyProfile()
	return cfg
}

// ApplyProfile expands a profile into concrete backend settings.
// profile: "speed" = all remote (DeepSeek/OpenAI), "balanced" = catalog remote + vector local, "local" = all local
func (c *Config) ApplyProfile() {
	switch c.Profile {
	case "speed":
		if c.Catalog.Backend == "" || c.Catalog.Backend == "local" {
			c.Catalog.Backend = "remote"
		}
		if c.Catalog.Endpoint == "http://localhost:1234/v1" {
			c.Catalog.Endpoint = "https://api.deepseek.com/v1"
		}
		if c.Catalog.Model == "qwen/qwen3.5-9b" {
			c.Catalog.Model = "deepseek-chat"
		}
		if c.Vector.Mode == "" || c.Vector.Mode == "local" {
			c.Vector.Mode = "remote"
		}
		if c.Vector.Endpoint == "http://localhost:1234/v1" {
			c.Vector.Endpoint = "https://api.openai.com/v1"
		}
		if c.Vector.Model == "text-embedding-baai-bge-m3-568m:2" {
			c.Vector.Model = "text-embedding-3-small"
		}
		c.Features.Vector = true

	case "balanced":
		if c.Catalog.Backend == "" || c.Catalog.Backend == "local" {
			c.Catalog.Backend = "remote"
		}
		if c.Catalog.Endpoint == "http://localhost:1234/v1" {
			c.Catalog.Endpoint = "https://api.deepseek.com/v1"
		}
		if c.Catalog.Model == "qwen/qwen3.5-9b" {
			c.Catalog.Model = "deepseek-chat"
		}
		// Vector stays local
		if c.Vector.Mode == "" {
			c.Vector.Mode = "remote"
		}
		c.Features.Vector = true

	case "local":
		fallthrough
	default:
		if c.Catalog.Backend == "" {
			c.Catalog.Backend = "local"
		}
		if c.Vector.Mode == "" {
			c.Vector.Mode = "local"
		}
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
