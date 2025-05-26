package get

import (
	"context"
	"fmt"
	"strings"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/root"
)

func Get(_ *root.Flags) *cli.Command {
	return cli.NewCommand("get", "Get a naisdevice setting.",
		cli.WithArgs("setting"),
		cli.WithRun(run),
		cli.WithValidate(cli.ValidateExactArgs(1)),
	)
}

func run(ctx context.Context, w output.Output, args []string) error {
	setting := args[0]

	values, err := get(ctx)
	if err != nil {
		return err
	}

	if strings.EqualFold(setting, "autoconnect") {
		w.Printf("%v:\t%v\n", setting, values.AutoConnect)
	} else if strings.EqualFold(setting, "iloveninetiesboybands") {
		w.Printf("%v:\t%v\n", setting, values.ILoveNinetiesBoybands)
	} else {
		return fmt.Errorf("unknown setting: %v", setting)
	}

	return nil
}
