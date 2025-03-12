package buildmeta

var (
	AppVersion  string
	BuildNumber string
)

// Use ldflags to set build metadata for apps using Nexus Core
// go build -ldflags "-X github.com/SneaksAndData/nexus-core/pkg/buildmeta.AppVersion=0.0.0 -X github.com/SneaksAndData/nexus-core/pkg/buildmeta.BuildNumber=dev1"
