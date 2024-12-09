package buildvars

// Build variables. Version, CommitID and BuildTime is overwritten during build time.
var (
	// Do not change Version, it is updated while releasing.
	Version = "3.16.0"
	// Do not change CommitID, it is updated while releasing.
	CommitID = "0000000000000000000000000000000000000000"
	// Do not change BuildTime, it is updated while releasing.
	BuildTime = "2021-01-01T00:00:00"
	// BuildMode is the mode of the build. It can be "development" or "production".
	BuildMode = "production"
)

func GetVersion() string {
	return Version
}

func GetCommitID() string {
	return CommitID
}

func GetBuildTime() string {
	return BuildTime
}

func GetBuildMode() string {
	if BuildMode == "" {
		return "production"
	}
	return BuildMode
}
