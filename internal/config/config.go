package config

import (
	"os"

	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
	"github.com/spf13/viper"
)

// envFallback reads an env var and sets it on viper if non-empty.
// This ensures env vars are picked up even without a config file.
func envFallback(key, envVar string) {
	if v := os.Getenv(envVar); v != "" {
		viper.Set(key, v)
	}
}

// Load reads baize.yaml and returns a Config.
// Environment variables override file values:
//
//	BAIZE_CATALOG_ENDPOINT   - Catalog remote API endpoint
//	BAIZE_CATALOG_MODEL      - Catalog remote model name
//	BAIZE_CATALOG_API_KEY    - Catalog remote API key
//	BAIZE_VECTOR_ENDPOINT    - Vector remote API endpoint
//	BAIZE_VECTOR_MODEL       - Vector remote model name
//	BAIZE_VECTOR_API_KEY     - Vector remote API key
//	BAIZE_CATALOG_BACKEND    - Catalog backend (local|remote)
//	BAIZE_VECTOR_MODE        - Vector mode (local|remote)
func Load(path string) (model.Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	// Env var overrides
	v.SetEnvPrefix("BAIZE")
	v.AutomaticEnv()

	cfg := model.DefaultConfig()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return cfg, err
		}
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, err
	}

	// Explicit env fallbacks (works without config file)
	envFallback("catalog.endpoint", "BAIZE_CATALOG_ENDPOINT")
	envFallback("catalog.model", "BAIZE_CATALOG_MODEL")
	envFallback("catalog.api_key", "BAIZE_CATALOG_API_KEY")
	envFallback("catalog.backend", "BAIZE_CATALOG_BACKEND")
	envFallback("vector.endpoint", "BAIZE_VECTOR_ENDPOINT")
	envFallback("vector.model", "BAIZE_VECTOR_MODEL")
	envFallback("vector.api_key", "BAIZE_VECTOR_API_KEY")
	envFallback("vector.mode", "BAIZE_VECTOR_MODE")

	// Re-unmarshal to pick up env var overrides
	_ = v.Unmarshal(&cfg)

	return cfg, nil
}
