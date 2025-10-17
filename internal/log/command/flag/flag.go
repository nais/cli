package flag

import (
	"time"

	"github.com/nais/cli/internal/alpha/command/flag"
)

type LogFlags struct {
	*flag.Alpha

	Environment    string        `name:"environment" short:"e" usage:"Filter logs to a specific |environment|."`
	Team           []string      `name:"team" short:"t" usage:"Filter logs to a single |team|. Can be repeated."`
	Workload       []string      `name:"workload" short:"w" usage:"Filter logs to a single |workload|. Can be repeated."`
	Container      []string      `name:"container" short:"c" usage:"Filter logs to a specific |container|. Can be repeated."`
	WithTimestamps bool          `name:"with-timestamps" usage:"Include timestamps in log output."`
	WithLabels     bool          `name:"with-labels" usage:"Include labels in log output."`
	RawQuery       string        `name:"raw-query" usage:"Provide a raw query to filter logs. See https://grafana.com/docs/loki/latest/logql/ for syntax."`
	Since          time.Duration `name:"since" short:"s" usage:"How far back in time to start the initial batch."`
	Limit          int           `name:"limit" short:"l" usage:"Maximum number of initial log lines."`
}
