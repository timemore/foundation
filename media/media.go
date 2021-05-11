package media

import "github.com/gabriel-vasile/mimetype"

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
