package command_test

import (
	"testing"

	"github.com/nais/cli/internal/log/command"
)

func TestQueryBuilder_Build(t *testing.T) {
	t.Run("nothing added", func(t *testing.T) {
		query := command.NewQueryBuilder().Build()

		if expected := `{service_name!=""}`; query != expected {
			t.Errorf("Expected query to be %q, got %q", expected, query)
		}
	})

	t.Run("multiple fields added", func(t *testing.T) {
		query := command.
			NewQueryBuilder().
			AddEnvironments("e1", "e2").
			AddTeams("t1", "t2").
			AddWorkloads("w1", "w2").
			AddContainers("c1", "c2").
			AddPods("p1", "p2").
			Build()

		expected := `{` +
			`service_name!=""` +
			`,k8s_cluster_name=~"e1|e2"` +
			`,service_namespace=~"t1|t2"` +
			`,service_name=~"w1|w2"` +
			`}` +
			` | k8s_container_name=~"c1|c2"` +
			` | k8s_pod_name=~"p1|p2"`

		if query != expected {
			t.Errorf("Expected query to be %q, got %q", expected, query)
		}
	})
}
