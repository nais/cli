package flag

import (
	alpha "github.com/nais/cli/internal/alpha/command/flag"
)

type Apply struct {
	*alpha.Alpha
	AllowIgnoredFields bool `name:"allow-ignored-fields" usage:"Warn instead of failing when a manifest contains fields that nais apply ignores (e.g. |metadata.namespace| or |metadata.annotations|)."`
}
