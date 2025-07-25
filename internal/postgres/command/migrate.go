package command

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/cli/internal/postgres/migrate/finalize"
	"github.com/nais/cli/internal/postgres/migrate/promote"
	"github.com/nais/cli/internal/postgres/migrate/rollback"
	"github.com/nais/cli/internal/postgres/migrate/setup"
	"github.com/nais/naistrix"
)

func migrateCommand(parentFlags *flag.Postgres) *naistrix.Command {
	flags := &flag.Migrate{Postgres: parentFlags}
	return &naistrix.Command{
		Name:        "migrate",
		Title:       "Migrate to a new SQL instance.",
		StickyFlags: flags,
		SubCommands: []*naistrix.Command{
			migrateSetupCommand(flags),
			migratePromoteCommand(flags),
			migrateFinalizeCommand(flags),
			migrateRollbackCommand(flags),
		},
	}
}

func migrateSetupCommand(parentFlags *flag.Migrate) *naistrix.Command {
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

	return &naistrix.Command{
		Name:        "setup",
		Title:       "Make necessary setup for a new SQL instance migration.",
		Description: "Setup will create a new (target) instance with updated configuration, and enable continuous replication of data from the source instance.",
		Args: []naistrix.Argument{
			{Name: "app_name"},
			{Name: "target_sql_instance_name"},
		},
		ValidateFunc: func(ctx context.Context, args []string) error {
			if flags.Tier != "" && !strings.HasPrefix(flags.Tier, "db-") {
				return fmt.Errorf("tier must start with `db-`")
			}

			if flags.InstanceType != "" && !strings.HasPrefix(flags.InstanceType, "POSTGRES_") {
				return fmt.Errorf("instance type must start with `POSTGRES_`")
			}

			return nil
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			return setup.Run(ctx, args[0], args[1], flags)
		},
	}
}

func migratePromoteCommand(parentFlags *flag.Migrate) *naistrix.Command {
	flags := &flag.MigratePromote{Migrate: parentFlags}
	return &naistrix.Command{
		Name:        "promote",
		Title:       "Promote the migrated instance to the new primary instance.",
		Description: "Promote will promote the target instance to the new primary instance, and update the application to use the new instance.",
		Flags:       flags,
		Args: []naistrix.Argument{
			{Name: "app_name"},
			{Name: "target_sql_instance_name"},
		},
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			return promote.Run(ctx, args[0], args[1], flags)
		},
	}
}

func migrateFinalizeCommand(parentFlags *flag.Migrate) *naistrix.Command {
	flags := &flag.MigrateFinalize{Migrate: parentFlags}
	return &naistrix.Command{
		Name:        "finalize",
		Title:       "Finalize the migration.",
		Description: "Finalize will remove the source instance and associated resources after a successful migration.",
		Args: []naistrix.Argument{
			{Name: "app_name"},
			{Name: "target_sql_instance_name"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			return finalize.Run(ctx, args[0], args[1], flags)
		},
	}
}

func migrateRollbackCommand(parentFlags *flag.Migrate) *naistrix.Command {
	flags := &flag.MigrateRollback{Migrate: parentFlags}
	return &naistrix.Command{
		Name:        "rollback",
		Title:       "Roll back the migration.",
		Description: "Rollback will roll back the migration, and restore the application to use the original instance.",
		Args: []naistrix.Argument{
			{Name: "app_name"},
			{Name: "target_sql_instance_name"},
		},
		Flags: flags,
		RunFunc: func(ctx context.Context, out naistrix.Output, args []string) error {
			return rollback.Run(ctx, args[0], args[1], flags)
		},
	}
}
