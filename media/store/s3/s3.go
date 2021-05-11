package s3

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/timemore/bootstrap/errors"
	mediastore "github.com/timemore/bootstrap/media/store"
)

type Config struct {
	Region          string `split_words:"true"`
	BucketName      string `split_words:"true"`
	AccessKeyID     string `split_words:"true"`
	SecretAccessKey string `split_words:"true"`
}

const ServiceName = "s3"

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
	if config == nil {
		return nil, errors.ArgMsg("config", "missing")
	}

	conf, ok := config.(*Config)
	if !ok {
		return nil, errors.ArgMsg("config", "type invalid")
	}
	if conf == nil || conf.Region == "" || conf.BucketName == "" {
		return nil, errors.ArgMsg("config", "field invalid")
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
