package command

import (
	"context"
	"fmt"

	"github.com/nais/cli/internal/alpha/postgres"
	"github.com/nais/cli/internal/alpha/postgres/command/flag"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/input"
	"github.com/nais/naistrix/output"
)

func instanceCreateCommand(parent *flag.Postgres) *naistrix.Command {
	f := &flag.InstanceCreate{Postgres: parent, Version: "18"}
	return &naistrix.Command{
		Name: "create", Title: "Create a Postgres instance.", Flags: f,
		Description: "Create a new Postgres instance. Provisioning continues asynchronously.",
		Args:        []naistrix.Argument{{Name: "postgres"}},
		Examples: []naistrix.Example{
			{Description: "Select the destination environment interactively and create a Postgres with platform defaults.", Command: "my-postgres -t my-team"},
			{Description: "Create a Postgres in a specific environment.", Command: "my-postgres -t my-team -e dev-gcp"},
			{Description: "Create a highly available Postgres with custom resources.", Command: "my-postgres -t my-team -e dev-gcp --high-availability --cpu 100m --memory 512Mi --disk-size 10Gi"},
		},
		ValidateFunc: validation.RequireTeam(f),
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			if f.Version != "18" {
				return fmt.Errorf("--version must be 18")
			}
			env, err := resolvePostgresCreateEnvironment(ctx, string(f.Environment))
			if err != nil {
				return err
			}
			name := args.Get("postgres")
			rows := instanceTarget(name, f.Team, env)
			rows = append(rows, []string{"Version", string(f.Version)}, []string{"High availability", fmt.Sprint(f.HighAvailability)})
			for _, setting := range []struct{ label, value string }{{"CPU", f.CPU}, {"Memory", f.Memory}, {"Disk size", f.DiskSize}} {
				if setting.value != "" {
					rows = append(rows, []string{setting.label, setting.value})
				}
			}
			out.Infoln("Create Postgres with this configuration (omitted resources use platform defaults):")
			if err := out.Table(output.TableWithMargins()).Render(rows); err != nil {
				return err
			}
			if !f.Yes {
				ok, err := input.Confirm("Create Postgres?", input.ConfirmWithDefaultTrue())
				if err != nil {
					return err
				}
				if !ok {
					return fmt.Errorf("cancelled by user")
				}
			}
			created, err := postgres.CreateInstance(ctx, gql.CreatePostgresInput{
				Name: name, TeamSlug: f.Team, EnvironmentName: env,
				MajorVersion: string(f.Version), HighAvailability: enabledHA(f.HighAvailability),
				Cpu: optional(f.CPU), Memory: optional(f.Memory), DiskSize: optional(f.DiskSize),
			})
			if err != nil {
				return err
			}
			printPostgresHeading(out, created, env)
			out.Println("Started creation of postgres. Run this command to check status:")
			out.Println(postgresGetCommandLine(created, f.Team, env, f.Config))
			return nil
		},
	}
}

func instanceUpdateCommand(parent *flag.Postgres) *naistrix.Command {
	f := &flag.InstanceUpdate{Postgres: parent}
	return &naistrix.Command{
		Name: "update", Title: "Update a Postgres instance.", Flags: f,
		Description: "Compare configured values with the requested changes before updating provided fields. Current resource values are configured requests, not effective runtime resources or SQL readiness. A subsequent nais apply may overwrite API updates on manifest-managed Postgres.",
		Args:        []naistrix.Argument{{Name: "postgres"}}, AutoCompleteFunc: autoCompletePostgresNames(parent),
		Examples: []naistrix.Example{
			{Description: "Change requested CPU and memory.", Command: "my-postgres -t my-team -e dev-gcp --cpu 200m --memory 1Gi"},
			{Description: "Disable high availability.", Command: "my-postgres -t my-team -e dev-gcp --high-availability false"},
		},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			if parent.Team == "" {
				return fmt.Errorf("missing required team, specify -t, --team or set a default team")
			}
			changes, err := updateFields(f)
			if err != nil {
				return err
			}
			name := args.Get("postgres")
			env, err := resolvePostgresEnvironment(ctx, parent.Team, name, string(parent.Environment))
			if err != nil {
				return err
			}
			current, err := postgres.GetPostgres(ctx, parent.Team, env, name)
			if err != nil {
				return err
			}
			rows := [][]string{{"Setting", "Current", "New"}}
			if changes.HighAvailability != nil {
				rows = append(rows, []string{"High availability", fmt.Sprint(current.HighAvailability), fmt.Sprint(*changes.HighAvailability)})
			}
			for _, setting := range []struct {
				label              string
				current, requested *string
			}{
				{"CPU", current.Resources.Cpu, changes.Cpu},
				{"Memory", current.Resources.Memory, changes.Memory},
				{"Disk size", current.Resources.DiskSize, changes.DiskSize},
			} {
				if setting.requested == nil {
					continue
				}
				value := "(not configured)"
				if setting.current != nil {
					value = *setting.current
				}
				rows = append(rows, []string{setting.label, value, *setting.requested})
			}
			out.Warnln("A subsequent nais apply can overwrite changes to manifest-managed Postgres.")
			if err := out.Table(output.TableWithMargins()).Render(instanceTarget(name, f.Team, env)); err != nil {
				return err
			}
			out.Println("Current values are configured settings, not effective runtime resources or SQL readiness.")
			if err := out.Table(output.TableWithMargins()).Render(rows); err != nil {
				return err
			}
			if !f.Yes {
				ok, err := input.Confirm("Request Postgres update?")
				if err != nil {
					return err
				}
				if !ok {
					return fmt.Errorf("cancelled by user")
				}
			}
			changes.Name, changes.TeamSlug, changes.EnvironmentName = name, f.Team, env
			updated, err := postgres.UpdateInstance(ctx, changes)
			if err != nil {
				return err
			}
			printPostgresHeading(out, updated, env)
			out.Println("Started updating Postgres. Run this command to check status:")
			out.Println(postgresGetCommandLine(updated, f.Team, env, f.Config))
			return nil
		},
	}
}

func instanceDeleteCommand(parent *flag.Postgres) *naistrix.Command {
	f := &flag.InstanceDelete{Postgres: parent}
	return &naistrix.Command{
		Name: "delete", Title: "Request deletion of an entire Postgres instance.", Flags: f,
		Description: "Irreversibly delete a Postgres and all its branches and data. Unlike branch delete, this removes the whole instance. Workloads and bindings must no longer reference it.",
		Args:        []naistrix.Argument{{Name: "postgres"}}, AutoCompleteFunc: autoCompletePostgresNames(parent),
		Examples: []naistrix.Example{{Description: "Request deletion of an entire Postgres after removing its consumers.", Command: "my-postgres -t my-team -e dev-gcp"}},
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			if parent.Team == "" {
				return fmt.Errorf("missing required team, specify -t, --team or set a default team")
			}
			name := args.Get("postgres")
			env, err := resolvePostgresEnvironment(ctx, parent.Team, name, string(parent.Environment))
			if err != nil {
				return err
			}
			out.Warnln("IRREVERSIBLE: deletion requests removal of ALL branches, PVCs/data, and archives/backups. The API refuses deletion while workloads or bindings reference this Postgres.")
			rows := append(instanceTarget(name, parent.Team, env), []string{"Impact", "All branches, data and backups"})
			if err := out.Table(output.TableWithMargins()).Render(rows); err != nil {
				return err
			}
			requested, err := confirmAndDeleteInstance(ctx, gql.DeletePostgresInput{
				Name: name, TeamSlug: parent.Team, EnvironmentName: env,
			}, f.Yes, func(prompt string) (bool, error) { return input.Confirm(prompt) }, postgres.DeleteInstance)
			if err != nil {
				return err
			}
			if !requested {
				return fmt.Errorf("Postgres %q deletion was not requested", name)
			}
			out.Printf("Started deleting Postgres %q in %q.\n", name, env)
			return nil
		},
	}
}

func instanceTarget(name, team, env string) [][]string {
	return [][]string{{"Field", "Value"}, {"Team", team}, {"Environment", env}, {"Name", name}}
}

func confirmAndDeleteInstance(ctx context.Context, target gql.DeletePostgresInput, yes bool, confirm func(string) (bool, error), remove func(context.Context, gql.DeletePostgresInput) (bool, error)) (bool, error) {
	if !yes {
		ok, err := confirm(fmt.Sprintf("Irreversibly delete Postgres %q for %q in %q and all its branches/data?", target.Name, target.TeamSlug, target.EnvironmentName))
		if err != nil {
			return false, err
		}
		if !ok {
			return false, fmt.Errorf("cancelled by user")
		}
	}
	return remove(ctx, target)
}

func enabledHA(enabled bool) *bool {
	if !enabled {
		return nil
	}
	return &enabled
}

func optional(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func updateFields(f *flag.InstanceUpdate) (gql.UpdatePostgresInput, error) {
	input := gql.UpdatePostgresInput{Cpu: optional(f.CPU), Memory: optional(f.Memory), DiskSize: optional(f.DiskSize)}
	switch f.HighAvailability {
	case "true":
		value := true
		input.HighAvailability = &value
	case "false":
		value := false
		input.HighAvailability = &value
	case "":
	default:
		return input, fmt.Errorf("--high-availability must be true or false")
	}
	if input.HighAvailability == nil && input.Cpu == nil && input.Memory == nil && input.DiskSize == nil {
		return input, fmt.Errorf("specify at least one of --high-availability, --cpu, --memory, or --disk-size")
	}
	return input, nil
}
