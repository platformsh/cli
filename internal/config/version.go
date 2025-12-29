package config

import (
	"runtime/debug"
	"sync"
)

// Version information, populated from build info or ldflags.
var (
	Version = "0.0.0"
	Commit  = "local"
	Date    = ""
	BuiltBy = "local"
)

var initOnce sync.Once

func init() {
	initOnce.Do(initVersion)
}

func initVersion() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	// Use module version if available and not overridden.
	if info.Main.Version != "" && info.Main.Version != "(devel)" && Version == "0.0.0" {
		Version = info.Main.Version
	}

	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			if Commit == "local" && setting.Value != "" {
				Commit = setting.Value
				if len(Commit) > 12 {
					Commit = Commit[:12]
				}
			}
		case "vcs.time":
			if Date == "" && setting.Value != "" {
				Date = setting.Value
			}
		case "vcs.modified":
			if setting.Value == "true" && Commit != "local" {
				Commit += "-dirty"
			}
		}
	}
}
