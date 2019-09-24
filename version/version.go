package version

import (
	"fmt"
	"runtime"
)

// AppName represents the application name as the 'user agent' on the larger Ethereum network.
const AppName = "Ethermint"

// Version contains the application semantic version.
//
// TODO: How do we want to version this being that an initial Ethermint has
// been developed?
const Version = "0.0.0"

// ProtocolVersion is the supported Ethereum protocol version (e.g., Homestead, Olympic, etc.)
const ProtocolVersion uint = 63

// GitCommit contains the git SHA1 short hash set by build flags.
var GitCommit = ""

// ClientVersion returns the full version string for identification on the larger Ethereum network.
func ClientVersion() string {
	return fmt.Sprintf("%s/%s+%s/%s/%s", AppName, Version, GitCommit, runtime.GOOS, runtime.Version())
}
