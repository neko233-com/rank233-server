package version

import (
	"fmt"
	"runtime"
)

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func Full() string {
	return fmt.Sprintf("%s (%s) built %s [go %s]", Version, Commit, Date, runtime.Version())
}
