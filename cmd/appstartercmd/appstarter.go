package appstartercmd

import (
	"fmt"

	"github.com/nais/cli/pkg/appstarter"
	"github.com/nais/cli/pkg/metrics"
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:      "start",
		Usage:     "Bootstrap basic yaml for nais and GitHub workflows",
		ArgsUsage: "teamname appname",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    "extra",
				Aliases: []string{"e"},
				Usage:   "list of desired extras (idporten,openSearch,aad,postgres), support repeating flags",
			},
			&cli.StringSliceFlag{
				Name:  "topic",
				Usage: "list of desired kafka topic resources, support repeating flags",
			},
			&cli.UintFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Usage:   "the port the app will listen on",
				Value:   8080,
			},
		},
		Before: func(context *cli.Context) error {
			metrics.AddOne("appstarter_total")
			if context.Args().Len() < 2 {
				metrics.AddOne("appstarter_arguments_error_total")
				return fmt.Errorf("missing required arguments: %v", context.Command.ArgsUsage)
			}

			return nil
		},
		Action: func(context *cli.Context) error {
			team := context.Args().Get(0)
			appName := context.Args().Get(1)

			extras := context.StringSlice("extras")
			topics := context.StringSlice("topics")
			port := context.Uint("port")

			return appstarter.Naisify(appName, team, extras, topics, port)
		},
	}
}
