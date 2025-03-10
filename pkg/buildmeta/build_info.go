package buildmeta

var (
	AppVersion  string
	BuildNumber string
)

// Use ldflags to set build metadata for apps using Nexus Core
// go build -ldflags "-X versioning.AppVersion=1.0.0 -X versioning.BuildNumber=12082019"
