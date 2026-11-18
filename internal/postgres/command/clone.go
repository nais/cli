package command

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/cli/internal/postgres"
	"github.com/nais/cli/internal/postgres/command/flag"
	"github.com/nais/cli/internal/validation"
	"github.com/nais/naistrix"
	"github.com/nais/naistrix/input"
	"github.com/nais/naistrix/output"
)

// cloneLatestBuffer keeps a default clone safely inside the archived WAL: the newest
// transactions may not have been archived yet.
const cloneLatestBuffer = 10 * time.Minute

func instanceCloneCommand(parent *flag.Postgres) *naistrix.Command {
	f := &flag.InstanceClone{Postgres: parent}
	return &naistrix.Command{
		Name: "clone", Title: "Clone a Postgres instance to a new database.", Flags: f,
		Description: "Create a new Postgres from the backups of an existing one. " +
			"Without --at the clone is made from the latest available point, approximately 10 minutes ago. " +
			"The new database is created in the same team and environment as the source, with the same PostgreSQL version and extensions. " +
			"The source is not changed and no applications are moved to the new database. " +
			"Availability and resources default to the source's configuration; the disk cannot be smaller than the source's. " +
			"Cloning continues asynchronously.",
		Args: []naistrix.Argument{
			{Name: "source", Prompt: "Name of the Postgres instance to clone", ChoicesFunc: postgresNameChoices(parent)},
			{Name: "postgres", Prompt: "Name of the new Postgres instance"},
		},
		Examples: []naistrix.CommandExample{
			{Description: "Clone a Postgres from the latest available point.", Command: "my-postgres my-clone -t my-team"},
			{Description: "Clone a Postgres as it was at a point in time (local time).", Command: "my-postgres my-clone --at \"2026-11-16 12:00\" -t my-team"},
			{Description: "Clone with more resources than the source.", Command: "my-postgres my-clone -t my-team -e dev-gcp --cpu 500m --memory 2Gi --disk-size 50Gi"},
		},
		ValidateFlagsFunc: validation.RequireTeamFlag(f),
		RunFunc: func(ctx context.Context, args *naistrix.Arguments, out *naistrix.OutputWriter) error {
			source, name := args.Get("source"), args.Get("postgres")
			if source == name {
				return fmt.Errorf("the new Postgres must have a different name than its source")
			}
			now := time.Now()
			target, explicit, err := resolveCloneTime(f.At, now, time.Local)
			if err != nil {
				return err
			}
			highAvailability, err := parseHighAvailability(f.HighAvailability)
			if err != nil {
				return err
			}
			env, err := resolvePostgresEnvironment(ctx, f.Team, source, string(f.Environment))
			if err != nil {
				return err
			}
			src, err := postgres.GetCloneSource(ctx, f.Team, env, source)
			if err != nil {
				return err
			}
			if err := validateCloneWindow(target, explicit, src.EarliestRestoreTime, time.Local); err != nil {
				return err
			}
			targetLabel := formatLocalTime(target, time.Local)
			if !explicit {
				targetLabel += " (latest available)"
			}
			rows := instanceTarget(name, f.Team, env)
			rows = append(rows,
				[]string{"Clone of", source},
				[]string{"Clone from", targetLabel},
				[]string{"Earliest possible", formatLocalTime(*src.EarliestRestoreTime, time.Local)},
				[]string{"Version", src.MajorVersion + " (same as source)"},
				[]string{"High availability", cloneSetting(fmt.Sprint(src.HighAvailability), highAvailabilityValue(highAvailability))},
				[]string{"CPU", cloneSetting(deref(src.Resources.Cpu), f.CPU)},
				[]string{"Memory", cloneSetting(deref(src.Resources.Memory), f.Memory)},
				[]string{"Disk size", cloneSetting(deref(src.Resources.DiskSize), f.DiskSize)},
			)
			out.Infoln("Clone Postgres with this configuration. The source is not changed and no applications are moved:")
			if err := out.Table(output.TableWithMargins()).Render(rows); err != nil {
				return err
			}
			if !f.Yes {
				ok, err := input.Confirm("Clone to a new Postgres?", input.ConfirmWithDefaultTrue())
				if err != nil {
					return err
				}
				if !ok {
					return fmt.Errorf("cancelled by user")
				}
			}
			created, err := postgres.CreateInstance(ctx, gql.CreatePostgresInput{
				Name: name, TeamSlug: f.Team, EnvironmentName: env, MajorVersion: src.MajorVersion,
				HighAvailability: highAvailability,
				Cpu:              optional(f.CPU), Memory: optional(f.Memory), DiskSize: optional(f.DiskSize),
				Recovery: &gql.PostgresRecoveryInput{SourcePostgres: source, TargetTime: target},
			})
			if err != nil {
				return err
			}
			printPostgresHeading(out, created, env)
			out.Println("Started cloning postgres. Run this command to check status:")
			out.Println(postgresGetCommandLine(created, f.Team, env, f.Config))
			return nil
		},
	}
}

// localTimeLayouts are accepted without an offset and read in the local time zone.
var localTimeLayouts = []string{"2006-01-02T15:04:05", "2006-01-02 15:04:05", "2006-01-02T15:04", "2006-01-02 15:04"}

// resolveCloneTime returns the absolute instant to clone from, in whole seconds. An empty value
// means the latest available point (now minus the safety buffer). A value with an offset or Z is
// absolute; a value without one is read in loc. Fractional seconds and the future are rejected.
func resolveCloneTime(value string, now time.Time, loc *time.Location) (target time.Time, explicit bool, err error) {
	if value == "" {
		return now.Add(-cloneLatestBuffer).Truncate(time.Second).UTC(), false, nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		parsed, err = parseLocalTime(value, loc)
		if err != nil {
			return time.Time{}, true, err
		}
	}
	if parsed.Nanosecond() != 0 {
		return time.Time{}, true, fmt.Errorf("--at must be in whole seconds")
	}
	if parsed.After(now) {
		return time.Time{}, true, fmt.Errorf("--at must not be in the future")
	}
	return parsed.UTC(), true, nil
}

func parseLocalTime(value string, loc *time.Location) (time.Time, error) {
	for _, layout := range localTimeLayouts {
		wall, err := time.Parse(layout, value)
		if err != nil {
			continue
		}
		instants := localInstants(wall, layout, loc)
		if len(instants) == 0 {
			return time.Time{}, fmt.Errorf("--at %q does not exist in time zone %s (daylight saving change)", value, loc)
		}
		// A repeated time occurs more than once; choose the earliest, the same as Console.
		return slices.MinFunc(instants, func(a, b time.Time) int { return a.Compare(b) }), nil
	}
	return time.Time{}, fmt.Errorf("--at must be a local time like \"2026-11-16 12:00\" or RFC 3339 like 2026-11-16T12:00:00Z")
}

// localInstants returns every instant at which loc's wall clock shows wall: none when the time is
// skipped by a clock change, two when it is repeated. It tries the zone's actual offsets around the
// date rather than assuming a clock change is one hour.
func localInstants(wall time.Time, layout string, loc *time.Location) []time.Time {
	offsets := map[int]bool{}
	for _, day := range []int{-1, 0, 1} {
		_, offset := wall.AddDate(0, 0, day).In(loc).Zone()
		offsets[offset] = true
	}
	var instants []time.Time
	for offset := range offsets {
		instant := wall.Add(-time.Duration(offset) * time.Second)
		if instant.In(loc).Format(layout) == wall.Format(layout) {
			instants = append(instants, instant)
		}
	}
	return instants
}

// validateCloneWindow rejects targets outside the backup window. A default target that is
// too early gets its own hint, since the user never chose it.
func validateCloneWindow(target time.Time, explicit bool, earliest *time.Time, loc *time.Location) error {
	if earliest == nil {
		return fmt.Errorf("the source has no recoverable backup window yet; try again later")
	}
	if target.Before(*earliest) {
		if !explicit {
			return fmt.Errorf("the source's backups only reach back to %s, which is later than the default clone point (about 10 minutes ago); try again shortly or specify --at", formatLocalTime(*earliest, loc))
		}
		return fmt.Errorf("--at is before the earliest possible clone time %s", formatLocalTime(*earliest, loc))
	}
	return nil
}

// formatLocalTime shows an instant in loc with its zone, so the output is unambiguous.
func formatLocalTime(t time.Time, loc *time.Location) string {
	return t.In(loc).Format("2006-01-02 15:04:05 MST")
}

func parseHighAvailability(value flag.HighAvailability) (*bool, error) {
	switch value {
	case "":
		return nil, nil
	case "true", "false":
		enabled := value == "true"
		return &enabled, nil
	}
	return nil, fmt.Errorf("--high-availability must be true or false")
}

func highAvailabilityValue(value *bool) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(*value)
}

// cloneSetting shows an explicit override, or the inherited source value.
func cloneSetting(inherited, explicit string) string {
	if explicit != "" {
		return explicit + " (overrides source)"
	}
	if inherited == "" {
		return "platform default"
	}
	return inherited + " (from source)"
}

func deref(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
