package store

import (
	"bytes"
	"io"
)

type ServiceConfig interface{}

type Service interface {
	PutObject(objectKey string, content io.Reader) (uploadInfo interface{}, err error)
	GetObject(objectKey string) (buffer *bytes.Buffer, err error)
	GetPublicObject(objectKey string) (string, error)
}
