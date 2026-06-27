package buildinfo

import (
	"runtime/debug"

	"github.com/lathe-cli/lathe/pkg/lathe"
)

const (
	defaultVersion = "dev"
	defaultCommit  = "none"
	defaultDate    = "unknown"
	develVersion   = "(devel)"
)

type Info struct {
	Version  string `json:"version"`
	Commit   string `json:"commit"`
	Date     string `json:"date"`
	Complete bool   `json:"complete"`
}

var readBuildInfo = debug.ReadBuildInfo

func Apply() {
	info := Current()
	lathe.Version = info.Version
	lathe.Commit = info.Commit
	lathe.Date = info.Date
}

func Current() Info {
	goInfo := currentGoBuildInfo()
	version := firstValue(defaultVersion, lathe.Version, goInfo.version)
	commit := firstValue(defaultCommit, lathe.Commit, goInfo.commit)
	date := firstValue(defaultDate, lathe.Date, goInfo.date)
	return Info{
		Version:  version,
		Commit:   commit,
		Date:     date,
		Complete: version != defaultVersion || (commit != defaultCommit && date != defaultDate),
	}
}

type goBuildInfo struct {
	version string
	commit  string
	date    string
}

func currentGoBuildInfo() goBuildInfo {
	info, ok := readBuildInfo()
	if !ok || info == nil {
		return goBuildInfo{}
	}
	return goBuildInfo{
		version: moduleVersion(info),
		commit:  settingValue(info, "vcs.revision"),
		date:    settingValue(info, "vcs.time"),
	}
}

func moduleVersion(info *debug.BuildInfo) string {
	if info.Main.Version == "" || info.Main.Version == develVersion {
		return ""
	}
	return info.Main.Version
}

func settingValue(info *debug.BuildInfo, key string) string {
	for _, setting := range info.Settings {
		if setting.Key == key {
			return setting.Value
		}
	}
	return ""
}

func firstValue(defaultValue string, values ...string) string {
	for _, value := range values {
		if value != "" && value != defaultValue {
			return value
		}
	}
	return defaultValue
}
