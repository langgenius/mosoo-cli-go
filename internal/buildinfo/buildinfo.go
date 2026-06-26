package buildinfo

const (
	defaultVersion = "dev"
	defaultCommit  = "none"
	defaultDate    = "unknown"
)

var (
	Version = defaultVersion
	Commit  = defaultCommit
	Date    = defaultDate
)

type Info struct {
	Version  string `json:"version"`
	Commit   string `json:"commit"`
	Date     string `json:"date"`
	Complete bool   `json:"complete"`
}

func Current() Info {
	version := valueOrDefault(Version, defaultVersion)
	commit := valueOrDefault(Commit, defaultCommit)
	date := valueOrDefault(Date, defaultDate)
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
