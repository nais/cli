package job

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
	"k8s.io/apimachinery/pkg/util/validation"
)

var invalidRunNameChars = regexp.MustCompile(`[^a-z0-9-]+`)

func TriggerJob(ctx context.Context, team, name, environment, runName string) (string, error) {
	_ = `# @genqlient
		mutation TriggerJob($team: Slug!, $name: String!, $env: String!, $runName: String!) {
			triggerJob(input: { teamSlug: $team, name: $name, environmentName: $env, runName: $runName }) {
				job {
					name
				}
				jobRun {
					name
				}
			}
		}
	`

	if environment == "" {
		return "", fmt.Errorf("exactly one environment must be specified")
	}
	env := environment

	if runName == "" {
		runName = autoGenerateRunName(name)
	}

	if err := validateRunName(runName); err != nil {
		return "", err
	}

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return "", err
	}

	resp, err := gql.TriggerJob(ctx, client, team, name, env, runName)
	if err != nil {
		return "", err
	}

	if resp.TriggerJob.JobRun.Name != "" {
		return fmt.Sprintf("Successfully triggered %s in %s (run: %s)", name, env, resp.TriggerJob.JobRun.Name), nil
	}

	return fmt.Sprintf("Successfully triggered %s in %s", name, env), nil
}

func autoGenerateRunName(jobName string) string {
	timestamp := time.Now().UTC().Format("20060102-150405")
	suffix := "-cli-" + timestamp

	base := strings.ToLower(jobName)
	base = invalidRunNameChars.ReplaceAllString(base, "-")
	base = strings.Trim(base, "-")
	if base == "" {
		base = "job"
	}

	maxBaseLen := max(validation.DNS1123LabelMaxLength-len(suffix), 1)
	if len(base) > maxBaseLen {
		base = strings.TrimRight(base[:maxBaseLen], "-")
		if base == "" {
			base = "job"
		}
	}

	return base + suffix
}

func validateRunName(runName string) error {
	if len(runName) > validation.DNS1123LabelMaxLength {
		return fmt.Errorf("run name is too long: max %d characters", validation.DNS1123LabelMaxLength)
	}
	if errs := validation.IsDNS1123Label(runName); len(errs) > 0 {
		return fmt.Errorf("invalid run name %q: %s", runName, strings.Join(errs, "; "))
	}
	return nil
}
