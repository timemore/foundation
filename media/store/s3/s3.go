package s3

import (
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/timemore/foundation/errors"
	mediastore "github.com/timemore/foundation/media/store"
)

type Config struct {
	Region          string `env:"REGION,required"`
	BucketName      string `env:"BUCKET_NAME,required"`
	AccessKeyID     string `env:"ACCESS_KEY_ID"`
	SecretAccessKey string `env:"SECRET_ACCESS_KEY"`
}

const ServiceName = "s3"

func init() {
	fmt.Printf("init service s3")
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
	if config == nil {
		return nil, errors.ArgMsg("config", "missing")
	}

	conf, ok := config.(*Config)
	if !ok {
		return nil, errors.ArgMsg("config", "type invalid")
	}
	if conf == nil || conf.Region == "" || conf.BucketName == "" {
		return nil, errors.ArgMsg("config", "fields invalid")
	}

	var creds *credentials.Credentials
	if conf.AccessKeyID != "" {
		creds = credentials.NewStaticCredentials(
			conf.AccessKeyID,
			conf.SecretAccessKey,
			"",
		)
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(conf.Region),
		Credentials: creds,
	})
	if err != nil {
		return nil, errors.Wrap("AWS Session", err)
	}

	const uploadPartSize = 10 * 1024 * 1024 // 10MiB

	return &Service{
		bucketName: conf.BucketName,
		uploader: s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
			u.PartSize = uploadPartSize
		}),
	}, nil
}

type Service struct {
	bucketName string
	uploader   *s3manager.Uploader
}

func (s *Service) PutObject(targetKey string, contentSource io.Reader) (publicURL string, err error) {
	result, err := s.uploader.Upload(&s3manager.UploadInput{
		Body:   contentSource,
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(targetKey),
	})
	if err != nil {
		return "", errors.Wrap("upload", err)
	}
	return result.Location, nil
}

var _ mediastore.Service = &Service{}
