package media

var fileAllowedContentTypes = []string{
	"application/msword",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	"application/pdf",
	"text/csv",
	"application/rtf",
	"text/plain",
}

type fileMediaTypeInfo struct {
	mediaType     MediaType
	directoryName string
}

func (typeInfo *fileMediaTypeInfo) MediaType() MediaType {
	if typeInfo.mediaType == MediaType_MEDIA_TYPE_UNSPECIFIED {
		return MediaType_FILE
	}
	return typeInfo.mediaType
}

func (typeInfo *fileMediaTypeInfo) DirectoryName() string {
	if typeInfo.directoryName == "" {
		panic("directory name is unspecified")
	}
	return typeInfo.directoryName
}

func (typeInfo *fileMediaTypeInfo) IsContentTypeAllowed(contentType string) bool {
	for _, ct := range fileAllowedContentTypes {
		if ct == contentType {
			return true
		}
	}
	return false
}
