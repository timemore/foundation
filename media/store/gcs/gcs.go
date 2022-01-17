package gcs

import (
	"bytes"
	"context"
	"io"
	"os"
	"path"
	"time"

	gcs "cloud.google.com/go/storage"
	"github.com/timemore/foundation/errors"
	mediastore "github.com/timemore/foundation/media/store"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type Config struct {
	BucketName     string `env:"BUCKET_NAME"`
	ProjectID      string `env:"PROJECT_ID"`
	CredentialFile string `env:"CREDENTIAL_FILE"`
	Basepath       string `env:"BASEPATH"`
}

const ServiceName = "gcs"

func init() {
	mediastore.RegisterModule(
		ServiceName,
		mediastore.Module{
			NewService: NewService,
			ServiceConfigSkeleton: func() mediastore.ServiceConfig {
				cfg := ConfigSkeleton()
				return &cfg
			},
		})
}

func ConfigSkeleton() Config { return Config{} }

func NewService(config mediastore.ServiceConfig) (mediastore.Service, error) {
	ctx := context.Background()
	if config == nil {
		return nil, errors.ArgMsg("config", "missing")
	}

	conf, ok := config.(*Config)
	if !ok {
		return nil, errors.ArgMsg("config", "type invalid")
	}

	isExists, err := conf.IsAvailableCredentials()
	if err != nil {
		return nil, errors.Wrap("credentialFile", err)
	}

	if !isExists {
		return nil, errors.ArgMsg("config", "missong credential file")
	}

	if conf.ProjectID == "" {
		return nil, errors.ArgMsg("config", "missiong project ID")
	}
	client, err := gcs.NewClient(ctx, option.WithCredentialsFile(conf.CredentialFile))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	bucketName := conf.BucketName
	bucket := client.Bucket(bucketName)
	it := client.Buckets(ctx, conf.ProjectID)
	for {
		battrs, err := it.Next()
		if err == iterator.Done {
			// Creates the new bucket.
			ctx, cancel := context.WithTimeout(ctx, time.Second*10)
			defer cancel()
			if err := bucket.Create(ctx, conf.ProjectID, nil); err != nil {
				return nil, errors.Wrap("bucket creation failed", err)
			}
			break
		}
		if err != nil {
			return nil, errors.Wrap("bucket iteration", err)
		}

		if battrs.Name == bucketName {
			break
		}
	}

	return &Service{
		bucketName: bucketName,
		projectID:  conf.ProjectID,
		gcsClient:  client,
		basePath:   conf.Basepath,
	}, nil
}

type Service struct {
	bucketName string
	projectID  string
	gcsClient  *gcs.Client
	basePath   string
}

type UploadInfo struct {
	Bucket string
	Key    string
}

func (s *Service) PutObject(targetKey string, contentSource io.Reader) (uploadInfo interface{}, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*50)
	defer cancel()
	if s.basePath != "" {
		targetKey = path.Join(s.basePath, targetKey)
	}
	// Upload an object with storage.Writer.
	wc := s.gcsClient.Bucket(s.bucketName).Object(targetKey).NewWriter(ctx)
	if _, err = io.Copy(wc, contentSource); err != nil {
		return nil, errors.Wrap("copy file io.Copy", err)
	}
	if err := wc.Close(); err != nil {
		return nil, errors.Wrap("writer.Close", err)
	}

	return UploadInfo{
		Bucket: s.bucketName,
		Key:    targetKey,
	}, nil
}

func (s *Service) GetPublicObject(sourceKey string) (targetURl string, err error) {
	return "", nil
}

func (s *Service) GetObject(sourceKey string) (stream *bytes.Buffer, err error) {
	return nil, nil
}

var _ mediastore.Service = &Service{}

func (conf *Config) IsAvailableCredentials() (bool, error) {
	if conf.CredentialFile == "" {
		return false, errors.ArgMsg("config.CredentialFile", "empty")
	}
	inf, err := os.Stat(conf.CredentialFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, errors.ArgMsg("config.CredentialFile", "notExists")
		}

		return false, errors.Wrap("credential file not valid", err)
	}

	return !inf.IsDir(), nil
}
