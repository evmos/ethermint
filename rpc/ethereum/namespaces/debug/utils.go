package debug

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime/pprof"
	"strings"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/tendermint/tendermint/libs/log"
)

// isCPUProfileConfigurationActivated checks if cpuprofile was configured via flag
func isCPUProfileConfigurationActivated(ctx *server.Context) bool {
	// TODO: use same constants as server/start.go
	// constant declared in start.go cannot be imported (cyclical dependency)
	const flagCPUProfile = "cpu-profile"
	if cpuProfile := ctx.Viper.GetString(flagCPUProfile); cpuProfile != "" {
		return true
	}
	return false
}

// ExpandHome expands home directory in file paths.
// ~someuser/tmp will not be expanded.
func ExpandHome(p string) string {
	if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") {
		home := os.Getenv("HOME")
		if home == "" {
			if usr, err := user.Current(); err == nil {
				home = usr.HomeDir
			}
		}
		if home != "" {
			p = home + p[1:]
		}
	}
	return filepath.Clean(p)
}

// writeProfile writes the data to a file
func writeProfile(name, file string, log log.Logger) error {
	p := pprof.Lookup(name)
	log.Info("Writing profile records", "count", p.Count(), "type", name, "dump", file)
	f, err := os.Create(ExpandHome(file))
	if err != nil {
		return err
	}
	defer f.Close()

	return p.WriteTo(f, 0)
}
