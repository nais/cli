package flag

import (
	"time"

	"github.com/nais/cli/internal/flags"
	"github.com/nais/naistrix"
)

type Apply struct {
	*flags.GlobalFlags
	AllowIgnoredFields bool                 `name:"allow-ignored-fields" usage:"Warn instead of failing when a manifest contains fields that nais apply ignores (e.g. |metadata.namespace| or |metadata.annotations|)."`
	DryRun             bool                 `name:"dry-run" usage:"Preview which resources would be applied without making any changes."`
	Mixin              mixinFile            `name:"mixin" usage:"YAML |FILE| deep-merged over the base manifest (mixin values win). If omitted, an adjacent <base>.<env>.yaml is auto-loaded when present."`
	Set                naistrix.StringArray `name:"set" usage:"Replace a field as |KEY=VALUE| using a dotted path with optional zero-based list indices (e.g. spec.env[0].value=debug). Values are parsed as YAML, including lists and maps. Can be repeated; applied in order."`
	Wait               bool                 `name:"wait" usage:"Wait for applied resources to become ready before returning. Currently supported for |Application| resources; other kinds are skipped."`
	Timeout            time.Duration        `name:"timeout" usage:"Maximum time to wait for resources to become ready when |--wait| is set. Examples: 30s, 5m, 10m."`
}

type mixinFile string

var _ naistrix.FileAutoCompleter = (*mixinFile)(nil)

func (mixinFile) FileExtensions() []string {
	return []string{"yaml", "yml"}
}
