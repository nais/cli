package resource

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/valkey"
)

func TestValkeySpecToValkey(t *testing.T) {
	for _, version := range []struct {
		raw  string
		want gql.ValkeyMajorVersion
	}{
		{raw: "8.1", want: "V8_1"},
		{raw: "9.0", want: "V9_0"},
		{raw: "9.1", want: "V9_1"},
		{raw: "10.2", want: "V10_2"},
	} {
		for _, persistence := range []struct {
			name string
			yaml string
			want *bool
		}{
			{name: "omitted"},
			{name: "empty", yaml: "\n  persistence: {}"},
			{name: "enabled", yaml: "\n  persistence:\n    disabled: false", want: new(false)},
			{name: "disabled", yaml: "\n  persistence:\n    disabled: true", want: new(true)},
		} {
			t.Run(version.raw+"/"+persistence.name, func(t *testing.T) {
				m, err := ParseManifest(mustDocument(t, fmt.Sprintf(`
version: v1
type: Valkey
name: my-cache
spec:
  memory: 4GB
  tier: SingleNode
  version: %q
  databases: 16
  maxMemoryPolicy: allkeys-lru
  notifyKeyspaceEvents: Ex
`, version.raw)+persistence.yaml))
				if err != nil {
					t.Fatal(err)
				}
				var spec valkeySpec
				if err := decodeSpec(&m.Spec, &spec); err != nil {
					t.Fatal(err)
				}
				data, err := spec.toValkey(map[string]string{"purpose": "cache"})
				if err != nil {
					t.Fatal(err)
				}
				want := &valkey.Valkey{
					Memory:               gql.ValkeyMemoryGb4,
					Tier:                 gql.ValkeyTierSingleNode,
					Version:              version.want,
					PersistenceDisabled:  persistence.want,
					Databases:            new(16),
					MaxMemoryPolicy:      gql.ValkeyMaxMemoryPolicyAllkeysLru,
					NotifyKeyspaceEvents: "Ex",
					Labels:               map[string]string{"purpose": "cache"},
				}
				if diff := cmp.Diff(want, data); diff != "" {
					t.Errorf("Valkey mismatch (-want +got):\n%s", diff)
				}
			})
		}
	}
}
