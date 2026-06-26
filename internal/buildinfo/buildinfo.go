package buildinfo

import "github.com/lathe-cli/lathe/pkg/lathe"

const (
	defaultVersion = "dev"
	defaultCommit  = "none"
	defaultDate    = "unknown"
)

type Info struct {
	Version  string `json:"version"`
	Commit   string `json:"commit"`
	Date     string `json:"date"`
	Complete bool   `json:"complete"`
}

func Current() Info {
	version := valueOrDefault(lathe.Version, defaultVersion)
	commit := valueOrDefault(lathe.Commit, defaultCommit)
	date := valueOrDefault(lathe.Date, defaultDate)
	return Info{
		Version:  version,
		Commit:   commit,
		Date:     date,
		Complete: version != defaultVersion && commit != defaultCommit && date != defaultDate,
	}
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
