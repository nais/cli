package appStarterCmd

import (
	"github.com/nais/cli/pkg/appStarter"
	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:        "start",
		Aliases:     []string{"v"},
		Description: "Bootstrap basic yaml for nais and GitHub workflows",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "appname",
				Aliases: []string{"n"},
				Usage:   "application name (as it will be in the cluster)",
			},
			&cli.StringFlag{
				Name:    "team",
				Aliases: []string{"t"},
				Usage:   "your team's name (app will be deployed to this namespace)",
			},
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
		HideHelpCommand: true,
		Action: func(context *cli.Context) error {
			appName := context.String("appname")
			team := context.String("team")
			extras := context.StringSlice("extras")
			topics := context.StringSlice("topics")
			port := context.Uint("port")

			return appStarter.Naisify(appName, team, extras, topics, port)
		},
	}
}
