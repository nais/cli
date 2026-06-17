package flag

import (
	"time"

	alpha "github.com/nais/cli/internal/alpha/command/flag"
	"github.com/nais/naistrix"
)

type Apply struct {
	*alpha.Alpha
	AllowIgnoredFields bool          `name:"allow-ignored-fields" usage:"Warn instead of failing when a manifest contains fields that nais apply ignores (e.g. |metadata.namespace| or |metadata.annotations|)."`
	Mixin              mixinFile     `name:"mixin" usage:"YAML |FILE| deep-merged over the base manifest (mixin values win). If omitted, an adjacent <base>.<env>.yaml is auto-loaded when present."`
	Set                []string      `name:"set" usage:"Override a single scalar field as |KEY=VALUE| using a dotted path (e.g. spec.image=ghcr.io/nais/app:latest). The value is parsed as YAML. Can be repeated."`
	Wait               bool          `name:"wait" usage:"Wait for applied resources to become ready before returning. Currently supported for |Application| resources; other kinds are skipped."`
	Timeout            time.Duration `name:"timeout" usage:"Maximum time to wait for resources to become ready when |--wait| is set. Examples: 30s, 5m, 10m."`
}

type mixinFile string

var _ naistrix.FileAutoCompleter = (*mixinFile)(nil)

func (mixinFile) FileExtensions() []string {
	return []string{"yaml", "yml"}
}
