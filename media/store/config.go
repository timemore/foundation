package store

import (
	"github.com/rez-go/stev"
)

type Config struct {
	NameGenerationKey string `env:"FILENAME_GENERATION_KEY"`
	StoreService      string `env:"STORE_SERVICE"`

	Modules map[string]any `env:",map,squash"`

	ImagesBaseURL string `env:"IMAGES_BASE_URL"`
}

// ParseConfigFromEnv populate the configuration by looking up the environment variables.
func ParseConfigFromEnv(prefix string) (*Config, error) {
	cfg := Config{}
	err := stev.LoadEnv(prefix, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
