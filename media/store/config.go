package store

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	NameGenerationKey string `split_words:"true" envconfig:"FILENAME_GENERATION_KEY"`
	StoreService      string `split_words:"true" envconfig:"STORE_SERVICE"`

	Modules map[string]interface{} `split_words:"true"`

	ImagesBaseURL string `envconfig:"IMAGES_BASE_URL" split_words:"true"`
}

func ParseConfigFromEnv(prefix string) (*Config, error) {
	cfg := Config{}
	err := envconfig.Process(prefix, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
