package get

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/root"
)

func Get(rootFlags *root.Flags) *cli.Command {
	return cli.NewCommand("get", "Get a naisdevice setting.",
		cli.WithPositionalArgs("setting"),
		cli.WithHandler(run),
		cli.WithExactArgs(1),
	)
}

func run(ctx context.Context, args []string) error {
	setting := args[0]

	values, err := get(ctx)
	if err != nil {
		return err
	}

	if strings.EqualFold(setting, "autoconnect") {
		fmt.Printf("%v:\t%v\n", setting, values.AutoConnect)
	} else if strings.EqualFold(setting, "iloveninetiesboybands") {
		fmt.Printf("%v:\t%v\n", setting, values.ILoveNinetiesBoybands)
	} else {
		return fmt.Errorf("unknown setting: %v", setting)
	}

	return nil
}
