package config

import (
	"github.com/kuaizhongqiang/baize-wiki/internal/core/model"
	"github.com/spf13/viper"
)

// Load reads baize.yaml and returns a Config.
// Returns default config if no file is found.
func Load(path string) (model.Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	// Env var overrides
	v.SetEnvPrefix("BAIZE")
	v.AutomaticEnv()

	cfg := model.DefaultConfig()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return cfg, nil
		}
		return cfg, err
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
