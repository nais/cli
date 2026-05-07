package flag

import (
	"context"

	activityutil "github.com/nais/cli/internal/activity"
	"github.com/nais/cli/internal/flags"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
)

type Config struct {
	*flags.GlobalFlags
}

type Output string

func (o *Output) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return []string{"table", "json"}, "Available output formats."
}

type List struct {
	*Config
	Output Output `name:"output" short:"o" usage:"Format output (table or json)."`
}

type Activity struct {
	*Config
	Output       Output        `name:"output" short:"o" usage:"Format output (table or json)."`
	Limit        int           `name:"limit" short:"l" usage:"Maximum number of activity entries to fetch."`
	ActivityType ActivityTypes `name:"activity-type" usage:"Filter by activity type. Can be repeated."`
}

type ActivityTypes []string

func (a *ActivityTypes) AutoComplete(context.Context, *naistrix.Arguments, string, any) ([]string, string) {
	return activityutil.EnumStrings(gql.AllActivityLogActivityType), "Available activity types"
}

type Get struct {
	*Config
	Output Output `name:"output" short:"o" usage:"Format output (table or json)."`
	ToFile string `name:"to-file" usage:"Write a single key's value to a file. Requires --key. Binary values are decoded automatically."`
	Key    string `name:"key" usage:"Name of the key to extract. Used with --to-file."`
}

type Create struct {
	*Config
}

type Delete struct {
	*Config
	Yes bool `name:"yes" short:"y" usage:"Automatic yes to prompts; assume 'yes' as answer to all prompts and run non-interactively."`
}

type Set struct {
	*Config
	Key            string `name:"key" usage:"Name of the key to set."`
	Value          string `name:"value" usage:"Value to set."`
	ValueFromStdin bool   `name:"value-from-stdin" usage:"Read value from stdin."`
	ValueFromFile  string `name:"value-from-file" usage:"Read value from file (e.g. keystore.p12, cert.pem). Binary files are automatically Base64-encoded."`
}

type Unset struct {
	*Config
	Key string `name:"key" usage:"Name of the key to unset."`
	Yes bool   `name:"yes" short:"y" usage:"Automatic yes to prompts; assume 'yes' as answer to all prompts and run non-interactively."`
}
