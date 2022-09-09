package media

import (
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

func DetectType(buf []byte) string {
	// Detect always returns valid MIME.
	return mimetype.Detect(buf).String()
}

func DetectExtension(buf []byte) string {
	return mimetype.Detect(buf).Extension()
}

func DetectMime(stream Reader) (*mimetype.MIME, error) {
	mimetype.SetLimit(0)
	return mimetype.DetectReader(stream)
}

func IsAllowedContentType(contentType string, allowedContentType []string) bool {
	if len(allowedContentType) != 0 {
		for _, ct := range allowedContentType {
			if ct == contentType {
				return true
			}
		}
	}

	return false
}

type MediaTypeInfo interface {
	// MediaType returns the type of media for this info
	MediaType() MediaType

	// DirectoryName return a string which usually used to construct path for storing the media files
	DirectoryName() string

	// IsContentTypeAllowed returns true if the provided content type string is allowed for the media type
	IsContentTypeAllowed(contentType string) bool

	// MimeType returns mimetype.MIME wich is mimetype info of the document,
	// accept Reader as input, read whole file instead of partial byte of the Reader
	DetectReader(r Reader) (*mimetype.MIME, error)
}

var mediaTypeRegistry = map[MediaType]MediaTypeInfo{
	MediaType_IMAGE: &imageMediaTypeInfo{
		mediaType:     MediaType_IMAGE,
		directoryName: "images",
	},
	MediaType_FILE: &fileMediaTypeInfo{
		mediaType:     MediaType_FILE,
		directoryName: "files",
	},
}

func GetMediaTypeInfoByTypeName(mediaTypeName string) MediaTypeInfo {
	if mediaTypeName == "" {
		return nil
	}

	mediaTypeName = strings.ToUpper(mediaTypeName)
	if v, ok := MediaType_Value[mediaTypeName]; ok {
		mediaType := MediaType(v)
		return GetMediaTypeInfo(mediaType)
	}

	return nil
}

func GetMediaTypeInfo(mediaType MediaType) MediaTypeInfo {
	return mediaTypeRegistry[mediaType]
}
