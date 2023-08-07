package rootcmd

import (
	"github.com/urfave/cli/v2"
)

func Commands() []*cli.Command {
	return []*cli.Command{
		loginCommand(),
	}
}
