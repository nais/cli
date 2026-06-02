package command

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/nais/cli/internal/app"
	"github.com/nais/cli/internal/app/command/flag"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/input"
	"golang.org/x/term"
)

type imageReleaseOption struct {
	image     string
	deployed  time.Time
	isCurrent bool
}

func (o imageReleaseOption) String() string {
	when := "unknown"
	if !o.deployed.IsZero() {
		when = o.deployed.Local().Format("2006-01-02 15:04")
	}

	if o.isCurrent {
		return fmt.Sprintf("%v (deployed %v, current)", o.image, when)
	}
	return fmt.Sprintf("%v (deployed %v)", o.image, when)
}

func setImage(parentFlags *flag.App) *naistrix.Command {
	flags := &flag.SetImage{
		App: parentFlags,
	}

	return &naistrix.Command{
		Name:  "image",
		Title: "Set the container image for an application.",
		Description: "Rolls the application back (or forward) to a previously deployed container image. " +
			"An interactive selector of previous releases is shown. Changes are temporary and will be overwritten on next deploy.",
		Examples: []naistrix.Example{
			{
				Description: "Select an image to roll back to interactively.",
				Command:     "my-app -e dev",
			},
		},
		Args: []naistrix.Argument{
			{Name: "name"},
		},
		Flags:            flags,
		AutoCompleteFunc: autoCompleteAppNames(parentFlags),
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			name := args.Get("name")

			out.Warnln(
				"This only changes the container image. Other changes made to the environment,",
				"such as environment variables, secrets or configuration, are not affected.",
			)

			environment, err := resolveAppEnvironment(ctx, out, flags.Team, name, string(flags.Environment), false)
			if err != nil {
				return err
			}

			image, err := selectImage(ctx, flags.Team, name, environment)
			if err != nil {
				return err
			}

			ret, err := app.SetImage(ctx, flags.Team, name, environment, image)
			if err != nil {
				return err
			}

			out.Println(ret)
			out.Warnln("Changes are temporary and will be overwritten on next deploy.")
			return nil
		},
	}
}

func selectImage(ctx context.Context, team, name, env string) (string, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) { // #nosec G115
		return "", fmt.Errorf("this command requires an interactive terminal to select an image. Please run this command in an interactive terminal")
	}

	images, err := app.GetApplicationImages(ctx, team, name, env)
	if err != nil {
		return "", err
	}

	if len(images.History) == 0 {
		return "", fmt.Errorf("no release history found for %q in %q", name, env)
	}

	options := make([]imageReleaseOption, 0, len(images.History))
	for _, release := range images.History {
		options = append(options, imageReleaseOption{
			image:     release.Image,
			deployed:  release.DeployedAt,
			isCurrent: release.Image == images.Current,
		})
	}

	selected, err := input.Select(fmt.Sprintf("Select an image for %v in %v", name, env), options)
	if err != nil {
		return "", fmt.Errorf("selecting image: %w", err)
	}

	return selected.image, nil
}
