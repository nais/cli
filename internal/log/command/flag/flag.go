package flag

import (
	"github.com/nais/cli/internal/alpha/command/flag"
)

type LogFlags struct {
	Team      string `short:"t" long:"team" description:"Team of the workload" required:"true"`
	Workload  string `short:"w" long:"workload" description:"Name of the workload to fetch logs from" required:"true"`
	Container string `short:"c" long:"container" description:"Name of the container to fetch logs from (if multiple containers exist)"`
	Follow    bool   `short:"f" long:"follow" description:"Follow the log output"`
	Lines     int    `short:"n" long:"lines" description:"Number of lines to show from the end of the logs" default:"100"`

	*flag.Alpha
}
