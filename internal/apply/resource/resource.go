// Package resource defines the nais resources that `nais apply` understands and
// the registry that routes manifests to them.
//
// Each resource lives in its own file and registers itself in init. Resources
// are matched on more than their kind, since several can share one (e.g.
// different Application apiVersions): each declares the manifests it handles via
// the Resource interface, and its capabilities via the optional Applier
// (nais-api mutation) and Waiter (--wait) interfaces.
package resource

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/nais/naistrix"
)

// Resource is a kind that `nais apply` knows about, scoped to the manifest
// versions it handles.
type Resource interface {
	Kind() string

	// APIVersion is the Kubernetes apiVersion to stamp on a CRD built from a
	// stripped manifest. Empty for mutation-only resources that never become CRDs.
	APIVersion() string

	// Supports reports whether the resource handles a manifest with the given
	// stripped version (native manifests) or apiVersion (CRDs); exactly one is set.
	Supports(strippedVersion, apiVersion string) bool
}

// Applier is implemented by resources with a dedicated nais-api mutation, using
// create-or-update semantics.
type Applier interface {
	Apply(ctx context.Context, meta Metadata, m Manifest) (Action, error)
}

// Waiter is implemented by resources that support --wait. since is the apply
// time, used to tell a fresh rollout from the resource's previous state.
type Waiter interface {
	Wait(ctx context.Context, team, environment, name string, since time.Time, out *naistrix.OutputWriter) error
}

// kindSupport implements the Kind, APIVersion and Supports methods of Resource.
// A resource embeds it to declare its kind, the stripped version it accepts (if
// any), and the apiVersion it maps to (if it can be a CRD).
type kindSupport struct {
	kind            string
	strippedVersion string
	apiVersion      string
}

func (k kindSupport) Kind() string       { return k.kind }
func (k kindSupport) APIVersion() string { return k.apiVersion }

func (k kindSupport) Supports(strippedVersion, apiVersion string) bool {
	switch {
	case strippedVersion != "":
		return k.strippedVersion != "" && k.strippedVersion == strippedVersion
	case apiVersion != "":
		return k.apiVersion != "" && k.apiVersion == apiVersion
	default:
		return false
	}
}

const (
	pollInterval = 5 * time.Second

	// graceWindow lets a no-op apply finish quickly while still giving the
	// controller time to start a rollout for a real change before we conclude the
	// app is already up to date.
	graceWindow = 19 * time.Second
)

// Metadata identifies where a resource is applied.
type Metadata struct {
	Name            string
	TeamSlug        string
	EnvironmentName string
	Labels          map[string]string
}

// Action describes what an apply did to a resource.
type Action string

const (
	ActionCreated Action = "created"
	ActionUpdated Action = "updated"
)

// registry holds every registered resource, grouped by kind.
var registry = map[string][]Resource{}

func register(r Resource) {
	registry[r.Kind()] = append(registry[r.Kind()], r)
}

// ForManifest returns the resource handling a stripped manifest, matched on kind
// and version.
func ForManifest(m Manifest) (Resource, bool) {
	return resolve(m.Kind, m.Version, "")
}

// ForCRD returns the resource handling a regular CRD, matched on kind and
// apiVersion. A mismatching apiVersion yields no resource, so callers never act
// on an unexpected type.
func ForCRD(apiVersion, kind string) (Resource, bool) {
	return resolve(kind, "", apiVersion)
}

func resolve(kind, strippedVersion, apiVersion string) (Resource, bool) {
	for _, r := range registry[kind] {
		if r.Supports(strippedVersion, apiVersion) {
			return r, true
		}
	}
	return nil, false
}

// enumValue maps a user-facing CRD value to its GraphQL enum, erroring with the
// allowed values when the input is unknown.
func enumValue[T ~string](field, raw string, table map[string]T) (T, error) {
	if v, ok := table[raw]; ok {
		return v, nil
	}
	var zero T
	return zero, fmt.Errorf("invalid %s %q (allowed: %s)", field, raw, strings.Join(sortedKeys(table), ", "))
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
