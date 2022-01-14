package store

import (
	"io"

	"github.com/timemore/foundation/media"
)

type ServiceConfig interface{}

type Service interface {
	PutObject(objectKey string, content io.Reader) (uploadInfo interface{}, err error)
	GetObject(objectKey string) (stream media.Reader, err error)
	GetPublicObject(objectKey string) (string, error)
}
