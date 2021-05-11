package app

type BuildInfo struct {
	RevisionID string
	Timestamp  string
}

var (
	buildRevisionID = "unknown"
	buildTimestamp  = "unknown"
)

func SetBuildInfo(revisionID string, timestamp string) {
	buildRevisionID = revisionID
	buildTimestamp = timestamp
}

func GetBuildInfo() BuildInfo {
	return BuildInfo{
		RevisionID: buildRevisionID,
		Timestamp:  buildTimestamp,
	}
}
