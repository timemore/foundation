package store

import (
	"io"
)

type ServiceConfig interface{}

type Service interface {
	PutObject(objectKey string, content io.Reader) (publicURL string, err error)
}
