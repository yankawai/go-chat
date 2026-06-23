package build

import "runtime"

var (
	Version = "dev"
	Commit  = "local"
	Date    = "unknown"
)

type Info struct {
	Service   string `json:"service"`
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"goVersion"`
}

func NewInfo(service string) Info {
	return Info{
		Service:   service,
		Version:   Version,
		Commit:    Commit,
		Date:      Date,
		GoVersion: runtime.Version(),
	}
}
