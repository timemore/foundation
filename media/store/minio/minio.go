package minio

import (
	"bytes"
	"context"
	"io"
	"net/url"
	"strconv"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/timemore/foundation/errors"
	"github.com/timemore/foundation/media"
	mediastore "github.com/timemore/foundation/media/store"
)

type Config struct {
	Region          string `env:"REGION"`
	BucketName      string `env:"BUCKET_NAME"`
	AccessKeyID     string `env:"ACCESS_KEY_ID,required"`
	SecretAccessKey string `env:"SECRET_ACCESS_KEY,required"`
	Endpoint        string `env:"ENDPOINT,required"`
	UseSSL          bool   `env:"USE_SSL"`
}

const ServiceName = "minio"

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
	if conf.Endpoint == "" {
		return nil, errors.ArgMsg("config.Endpoint", "empty")
	}
	if conf.AccessKeyID == "" || conf.SecretAccessKey == "" {
		return nil, errors.ArgMsg("config", "access key required")
	}

	// 	Initialize minio client object
	var cred *credentials.Credentials
	cred = credentials.NewStaticV4(
		conf.AccessKeyID,
		conf.SecretAccessKey,
		"",
	)
	minioClient, err := minio.New(conf.Endpoint, &minio.Options{
		Creds:  cred,
		Secure: conf.UseSSL})

	if err != nil {
		return nil, errors.Wrap("minio client initialization", err)
	}

	// 	make bucket
	bucketName := conf.BucketName
	location := conf.Region
	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists != nil && !exists {
			return nil, errors.Wrap("bucket creation", err)
		}
	}

	return &Service{
		bucketName:  bucketName,
		minioClient: minioClient,
	}, nil
}

type Service struct {
	bucketName  string
	minioClient *minio.Client
}

func (s *Service) PutObject(targetKey string, contentSource io.Reader) (uploadInfo interface{}, err error) {
	ctx := context.Background()
	bucketName := s.bucketName
	info, err := s.minioClient.PutObject(ctx, bucketName, targetKey, contentSource, -1, minio.PutObjectOptions{})
	if err != nil {
		return nil, errors.Wrap("upload", err)
	}

	return info, nil
}

func (s *Service) GetPublicObject(sourceKey string) (targetURl string, err error) {
	ctx := context.Background()
	object, err := s.minioClient.GetObject(ctx, s.bucketName, sourceKey, minio.GetObjectOptions{})
	if err != nil {
		return "", err
	}

	objectBuf := new(bytes.Buffer)
	_, _ = objectBuf.ReadFrom(object)
	objectExt := media.DetectExtension(objectBuf.Bytes())
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", "attachment;filename="+strconv.Quote(sourceKey+objectExt))

	// Generates a presigned url which expires in a day.
	preSignedURL, err := s.minioClient.PresignedGetObject(context.Background(), s.bucketName, sourceKey, time.Second*24*60*60, reqParams)
	if err != nil {
		return "", err
	}
	targetURl = preSignedURL.String()
	return
}

var _ mediastore.Service = &Service{}
