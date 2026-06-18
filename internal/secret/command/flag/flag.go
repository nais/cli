package flag

import (
	"context"

	activityutil "github.com/nais/cli/internal/activity"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/labels"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
)

type Secret struct {
	*flags.GlobalFlags
}

type Output string

func (o *Output) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"table", "json"}, "Available output formats."
}

type List struct {
	*Secret
	Output Output              `name:"output" short:"o" usage:"Format output (table or json)."`
	Labels labels.LabelFilters `name:"label" short:"l" usage:"Filter by label in |KEY=VALUE| form. Can be repeated."`
}

func (*List) LabelFacetResource() string { return "secrets" }

type Activity struct {
	*Secret
	Output       Output        `name:"output" short:"o" usage:"Format output (table or json)."`
	Limit        int           `name:"limit" short:"l" usage:"Maximum number of activity entries to fetch."`
	ActivityType ActivityTypes `name:"activity-type" usage:"Filter by activity type. Can be repeated."`
}

type ActivityTypes []string

func (a *ActivityTypes) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return activityutil.EnumStrings(gql.AllActivityLogActivityType), "Available activity types"
}

type Get struct {
	*Secret
	Output     Output `name:"output" short:"o" usage:"Format output (table or json)."`
	WithValues bool   `name:"with-values" usage:"Also fetch and display secret values (access is logged)."`
	Reason     string `name:"reason" usage:"Reason for accessing secret values (min 10 chars). Used with --with-values."`
	ToFile     string `name:"to-file" usage:"Write a single key's value to a file (implies --with-values). Requires --key. Binary values are decoded automatically."`
	Key        string `name:"key" usage:"Name of the key to extract. Used with --to-file."`
}

type Create struct {
	*Secret
}

type Delete struct {
	*Secret
	Yes bool `name:"yes" short:"y" usage:"Automatic yes to prompts; assume 'yes' as answer to all prompts and run non-interactively."`
}

type Set struct {
	*Secret
	Key            string `name:"key" usage:"Name of the key to set."`
	Value          string `name:"value" usage:"Value to set."`
	ValueFromStdin bool   `name:"value-from-stdin" usage:"Read value from stdin."`
	ValueFromFile  string `name:"value-from-file" usage:"Read binary value from file (e.g. keystore.p12, cert.pem). The value is sent as BASE64-encoded."`
}

type Unset struct {
	*Secret
	Key string `name:"key" usage:"Name of the key to unset."`
	Yes bool   `name:"yes" short:"y" usage:"Automatic yes to prompts; assume 'yes' as answer to all prompts and run non-interactively."`
}
