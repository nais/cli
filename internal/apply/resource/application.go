package resource

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nais/cli/internal/naisapi"
	"github.com/nais/cli/internal/naisapi/gql"
	"github.com/nais/naistrix"
)

func init() {
	register(applicationResource{kindSupport{kind: "Application", apiVersion: "nais.io/v1alpha1"}})
}

// applicationResource waits for Application rollouts to become ready.
// Applications have no nais-api mutation yet, so they are applied as CRDs
// (apiVersion nais.io/v1alpha1) and this resource only implements Waiter.
type applicationResource struct{ kindSupport }

// instanceGroup is a set of pods sharing one configuration (image, env, mounted
// files). A healthy app that is not mid-rollout has exactly one group with all
// instances ready.
type instanceGroup struct {
	created   time.Time
	ready     int
	desired   int
	instances []instance
}

// instance carries the status fields used to explain why a pod is not yet ready.
type instance struct {
	name       string
	state      gql.ApplicationInstanceState
	ready      bool
	restarts   int
	message    string
	exitReason string
	exitCode   *int
}

// Wait polls an application until its rollout has converged to a single,
// fully-ready instance group. Since the apply endpoint returns no rollout
// handle, correlation is best-effort: we wait for a new instance group (created
// at/after the apply) to converge, but treat an already-healthy app as a no-op
// once graceWindow passes without a new rollout appearing.
func (a applicationResource) Wait(ctx context.Context, team, environment, name string, since time.Time, out *naistrix.OutputWriter) error {
	out.Printf("Application/%s: waiting to become ready...\n", name)

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	sawRollout := false
	lastSummary := ""
	printedStates := map[string]gql.ApplicationInstanceState{}
	var lastGroups []instanceGroup

	for {
		state, groups, err := a.getApplicationStatus(ctx, team, environment, name)
		switch {
		case err == nil:
			lastGroups = groups
			converged := a.isConverged(state, groups)

			a.reportProgress(out, name, state, groups, &lastSummary, printedStates)

			if a.rolloutObserved(groups, since) {
				sawRollout = true
			}

			if converged {
				if sawRollout {
					out.Successf("Application/%s: ready\n", name)
					return nil
				}
				if time.Since(since) >= graceWindow {
					out.Successf("Application/%s: already up to date\n", name)
					return nil
				}
			}
		case naisapi.IsNotFound(err):
			// The application may not exist in the API yet; keep polling.
		case ctx.Err() != nil:
			return a.waitError(name, lastGroups, ctx.Err())
		default:
			return fmt.Errorf("failed to read status for Application/%s: %w", name, err)
		}

		select {
		case <-ctx.Done():
			return a.waitError(name, lastGroups, ctx.Err())
		case <-ticker.C:
		}
	}
}

// isConverged reports whether the app has settled on a single fully-ready
// instance group with no rollout in progress.
func (a applicationResource) isConverged(state gql.ApplicationState, groups []instanceGroup) bool {
	if state != gql.ApplicationStateRunning {
		return false
	}
	if len(groups) != 1 {
		// Zero groups: nothing running yet. More than one: rollout mid-flight.
		return false
	}
	g := groups[0]
	return g.desired > 0 && g.ready == g.desired
}

// rolloutObserved reports whether a rollout from this apply has appeared: a new
// instance group created at/after the apply, or multiple groups coexisting.
func (a applicationResource) rolloutObserved(groups []instanceGroup, since time.Time) bool {
	if len(groups) > 1 {
		return true
	}
	for _, g := range groups {
		if !g.created.Before(since) {
			return true
		}
	}
	return false
}

// reportProgress prints the summary line when it changes and an instance's
// detail line only when that instance reaches a new state, so a status message
// that flaps between equivalent phrasings does not spam the output.
func (a applicationResource) reportProgress(out *naistrix.OutputWriter, name string, state gql.ApplicationState, groups []instanceGroup, lastSummary *string, printedStates map[string]gql.ApplicationInstanceState) {
	if summary := a.summaryLine(state, groups); summary != *lastSummary {
		out.Printf("Application/%s: %s\n", name, summary)
		*lastSummary = summary
	}

	for _, ins := range a.unhealthyInstances(groups) {
		if printedStates[ins.name] == ins.state {
			continue
		}
		out.Printf("Application/%s:   %s\n", name, a.instanceDetail(ins))
		printedStates[ins.name] = ins.state
	}
}

func (a applicationResource) summaryLine(state gql.ApplicationState, groups []instanceGroup) string {
	ready, desired := 0, 0
	for _, g := range groups {
		ready += g.ready
		desired += g.desired
	}
	return fmt.Sprintf("state=%s, instances ready %d/%d, groups=%d", state, ready, desired, len(groups))
}

func (a applicationResource) unhealthyInstances(groups []instanceGroup) []instance {
	var unhealthy []instance
	for _, g := range groups {
		for _, ins := range g.instances {
			if !ins.ready {
				unhealthy = append(unhealthy, ins)
			}
		}
	}
	return unhealthy
}

func (a applicationResource) instanceDetail(ins instance) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s (state %s", ins.name, ins.state)
	if ins.restarts > 0 {
		fmt.Fprintf(&b, ", restarts %d", ins.restarts)
	}
	b.WriteString(")")

	if reason := a.instanceReason(ins); reason != "" {
		b.WriteString(": ")
		b.WriteString(reason)
	}
	return b.String()
}

// instanceReason picks the most informative explanation for an unready instance:
// the status message, else the last exit reason and code.
func (a applicationResource) instanceReason(ins instance) string {
	if ins.message != "" {
		return ins.message
	}
	if ins.exitReason != "" {
		if ins.exitCode != nil {
			return fmt.Sprintf("%s (exit code %d)", ins.exitReason, *ins.exitCode)
		}
		return ins.exitReason
	}
	if ins.exitCode != nil {
		return fmt.Sprintf("exit code %d", *ins.exitCode)
	}
	return ""
}

// waitError reports why the wait ended before the app became ready, listing the
// last seen reason each unhealthy instance failed. The phrasing distinguishes a
// timeout from a cancellation (e.g. a user interrupt).
func (a applicationResource) waitError(name string, groups []instanceGroup, cause error) error {
	prefix := "timed out waiting for"
	if errors.Is(cause, context.Canceled) {
		prefix = "canceled while waiting for"
	}

	var details strings.Builder
	if unhealthy := a.unhealthyInstances(groups); len(unhealthy) > 0 {
		details.WriteString("; unhealthy instances:")
		for _, ins := range unhealthy {
			details.WriteString("\n    ")
			details.WriteString(a.instanceDetail(ins))
		}
	}

	return fmt.Errorf("%s Application/%s to become ready (%w)%s", prefix, name, cause, details.String())
}

func (a applicationResource) getApplicationStatus(ctx context.Context, team, environment, name string) (gql.ApplicationState, []instanceGroup, error) {
	_ = `# @genqlient
		query ApplicationStatus($team: Slug!, $environment: String!, $name: String!) {
		  team(slug: $team) {
			environment(name: $environment) {
			  application(name: $name) {
				state
				instanceGroups {
				  name
				  created
				  readyInstances
				  desiredInstances
				  instances {
					name
					restarts
					status {
					  state
					  ready
					  message
					  # @genqlient(pointer: true)
					  lastExitReason
					  # @genqlient(pointer: true)
					  lastExitCode
					}
				  }
				}
			  }
			}
		  }
		}
	`

	client, err := naisapi.GraphqlClient(ctx)
	if err != nil {
		return "", nil, err
	}

	resp, err := gql.ApplicationStatus(ctx, client, team, environment, name)
	if err != nil {
		return "", nil, err
	}

	app := resp.Team.Environment.Application
	groups := make([]instanceGroup, 0, len(app.InstanceGroups))
	for _, g := range app.InstanceGroups {
		instances := make([]instance, 0, len(g.Instances))
		for _, ins := range g.Instances {
			instances = append(instances, instance{
				name:       ins.Name,
				state:      ins.Status.State,
				ready:      ins.Status.Ready,
				restarts:   ins.Restarts,
				message:    ins.Status.Message,
				exitReason: derefString(ins.Status.LastExitReason),
				exitCode:   ins.Status.LastExitCode,
			})
		}
		groups = append(groups, instanceGroup{
			created:   g.Created,
			ready:     g.ReadyInstances,
			desired:   g.DesiredInstances,
			instances: instances,
		})
	}
	return app.State, groups, nil
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
