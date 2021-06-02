package local

import (
	"io"
	"os"
	"path/filepath"

	"github.com/timemore/foundation/errors"
	mediastore "github.com/timemore/foundation/media/store"
)

type Config struct {
	DirectoryPath string `env:"FOLDER_PATH"`
}

const ServiceName = "local"

func init() {
	mediastore.RegisterModule(
		ServiceName,
		mediastore.Module{
			NewService: NewService,
			ServiceConfigSkeleton: func() mediastore.ServiceConfig {
				cfg := ConfigSkeleton()
				return cfg
			},
		})
}

func NewService(config mediastore.ServiceConfig) (mediastore.Service, error) {
	if config == nil {
		return nil, errors.ArgMsg("config", "missing")
	}

	conf, ok := config.(*Config)
	if !ok {
		return nil, errors.ArgMsg("config", "type invalid")
	}

	return &Service{
		directoryPath: conf.DirectoryPath,
	}, nil
}

type Service struct {
	directoryPath string
}

func (s *Service) PutObject(objectKey string, contentSource io.Reader) (uploadInfo interface{}, err error) {
	targetName := filepath.Join(s.directoryPath, objectKey)
	targetFile, err := os.Create(targetName)
	if err != nil {
		return "", errors.Wrap("create file", err)
	}
	defer func() {
		_ = targetFile.Close()
	}()

	_, err = io.Copy(targetFile, contentSource)
	if err != nil {
		return "", errors.Wrap("write content", err)
	}

	return UploadOutput{
		Directory: s.directoryPath,
		Key:       objectKey,
	}, nil
}

func (s *Service) GetPublicObject(sourceKey string) (targetURL string, err error) {
	// TODO: final URL! ask the HTTP server to provide this.
	return
}

var _ mediastore.Service = &Service{}

func ConfigSkeleton() Config { return Config{} }

type UploadOutput struct {
	Directory string
	Key       string
}
