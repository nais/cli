package command_test

import (
	"fmt"
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
			AddTeam("t1").AddTeam("t2").
			AddEnvironment("e1").AddEnvironment("e2").
			AddWorkload("w1").AddWorkload("w2").
			AddContainer("c1").AddContainer("c2").
			Build()

		expected := `{` +
			`service_name!=""` +
			`,service_namespace=~"t1|t2"` +
			`,k8s_cluster_name=~"e1|e2"` +
			`,service_name=~"w1|w2"` +
			`}` +
			` | k8s_container_name=~"c1|c2"`

		if query != expected {
			t.Errorf("Expected query to be %q, got %q", expected, query)
		}

		fmt.Println(expected)
	})
}
