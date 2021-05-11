package media

import (
	"strings"

	"github.com/gabriel-vasile/mimetype"
)

func DetectType(buf []byte) string {
	// Detect always returns valid MIME.
	return mimetype.Detect(buf).String()
}

type MediaTypeInfo interface {
	// MediaType returns the type of media for this info
	MediaType() MediaType

	// DirectoryName return a string which usually used to construct path for storing the media files
	DirectoryName() string

	// IsContentTypeAllowed returns true if the provided content type string is allowed for the media type
	IsContentTypeAllowed(contentType string) bool
}

var mediaTypeRegistry = map[MediaType]MediaTypeInfo{
	MediaType_IMAGE: &imageMediaTypeInfo{
		mediaType:     MediaType_IMAGE,
		directoryName: "images",
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
