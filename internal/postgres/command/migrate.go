package command

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/nais/cli/internal/cli"
	"github.com/nais/cli/internal/output"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/cli/internal/postgres/migrate"
	"github.com/nais/cli/internal/postgres/migrate/finalize"
	"github.com/nais/cli/internal/postgres/migrate/promote"
	"github.com/nais/cli/internal/postgres/migrate/rollback"
	"github.com/nais/cli/internal/postgres/migrate/setup"
)

func migrateCommand(parentFlags *flag.Postgres) *cli.Command {
	flags := &flag.Migrate{Postgres: parentFlags}

	return cli.NewCommand("migrate", "Migrate to a new SQL instance.",
		cli.WithStickyFlag("dry-run", "", "Perform a dry run.", &flags.DryRun),
		cli.WithSubCommands(
			migrateSetupCommand(flags),
			migratePromoteCommand(flags),
			migrateFinalizeCommand(flags),
			migrateRollbackCommand(flags),
		),
	)
}

func migrateSetupCommand(parentFlags *flag.Migrate) *cli.Command {
	flags := &flag.MigrateSetup{
		Migrate: parentFlags,
		Tier:    os.Getenv("TARGET_INSTANCE_TIER"),
	}

	if v, err := strconv.ParseBool(os.Getenv("TARGET_INSTANCE_DISK_AUTORESIZE")); err == nil {
		flags.DiskAutoResize = v
	}

	if v, err := strconv.Atoi(os.Getenv("TARGET_INSTANCE_DISKSIZE")); err == nil {
		flags.DiskSize = v
	}

	return cli.NewCommand("setup", "Make necessary setup for a new SQL instance migration.",
		cli.WithLong("Setup will create a new (target) instance with updated configuration, and enable continuous replication of data from the source instance."),
		cli.WithArgs("app_name", "target_sql_instance_name"),
		cli.WithValidate(
			cli.ValidateExactArgs(2),
			func(ctx context.Context, args []string) error {
				if !strings.HasPrefix(flags.Tier, "db-") {
					return fmt.Errorf("tier must start with `db-`")
				}

				if !strings.HasPrefix(flags.InstanceType, "POSTGRES_") {
					return fmt.Errorf("instance type must start with `POSTGRES_`")
				}

				return nil
			},
		),
		cli.WithRun(func(ctx context.Context, output output.Output, args []string) error {
			return setup.Run(
				ctx,
				migrate.Arguments{
					ApplicationName:    args[0],
					TargetInstanceName: args[1],
				},
				flags,
			)
		}),
		cli.WithFlag("no-wait", "", "Do not wait for the job to complete.", &flags.NoWait),
		cli.WithFlag("tier", "", "The `TIER` of the new instance.", &flags.Tier, cli.FlagRequired()),
		cli.WithFlag("disk-autoresize", "", "Enable disk autoresize for the new instance.", &flags.DiskAutoResize, cli.FlagRequired()),
		cli.WithFlag("disk-size", "", "The `DISK_SIZE` of the new instance.", &flags.DiskSize, cli.FlagRequired()),
		cli.WithFlag("type", "", "The `TYPE` of the new instance.", &flags.InstanceType, cli.FlagRequired()),
	)
}

func migratePromoteCommand(parentFlags *flag.Migrate) *cli.Command {
	flags := &flag.MigratePromote{
		Migrate: parentFlags,
	}

	return cli.NewCommand("promote", "Promote the migrated instance to the new primary instance.",
		cli.WithLong("Promote will promote the target instance to the new primary instance, and update the application to use the new instance."),
		cli.WithArgs("app_name", "target_sql_instance_name"),
		cli.WithValidate(cli.ValidateExactArgs(2)),
		cli.WithRun(func(ctx context.Context, output output.Output, args []string) error {
			return promote.Run(
				ctx,
				migrate.Arguments{
					ApplicationName:    args[0],
					TargetInstanceName: args[1],
				},
				flags,
			)
		}),
		cli.WithFlag("no-wait", "", "Do not wait for the job to complete.", &flags.NoWait),
	)
}

func migrateFinalizeCommand(parentFlags *flag.Migrate) *cli.Command {
	return cli.NewCommand("finalize", "Finalize the migration.",
		cli.WithLong("Finalize will remove the source instance and associated resources after a successful migration."),
		cli.WithArgs("app_name", "target_sql_instance_name"),
		cli.WithValidate(cli.ValidateExactArgs(2)),
		cli.WithRun(func(ctx context.Context, output output.Output, args []string) error {
			return finalize.Run(
				ctx,
				migrate.Arguments{
					ApplicationName:    args[0],
					TargetInstanceName: args[1],
				},
				&flag.MigrateFinalize{
					Migrate: parentFlags,
				},
			)
		}),
	)
}

func migrateRollbackCommand(parentFlags *flag.Migrate) *cli.Command {
	return cli.NewCommand("rollback", "Roll back the migration.",
		cli.WithLong("Rollback will roll back the migration, and restore the application to use the original instance."),
		cli.WithArgs("app_name", "target_sql_instance_name"),
		cli.WithValidate(cli.ValidateExactArgs(2)),
		cli.WithRun(func(ctx context.Context, output output.Output, args []string) error {
			return rollback.Run(
				ctx,
				migrate.Arguments{
					ApplicationName:    args[0],
					TargetInstanceName: args[1],
				},
				&flag.MigrateRollback{
					Migrate: parentFlags,
				},
			)
		}),
	)
}
