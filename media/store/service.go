package store

import (
	"io"
)

type ServiceConfig interface{}

type Service interface {
	PutObject(objectKey string, content io.Reader) (uploadInfo interface{}, err error)
	GetPublicObject(objectKey string) (string, error)
}
