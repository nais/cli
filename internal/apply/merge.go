package apply

import (
	"fmt"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// deepMerge merges override into base and returns the result. Semantics:
//   - two maps are merged recursively, key by key;
//   - two lists are concatenated (base elements first, then override);
//   - anything else: override wins (scalars and type changes replace base).
//
// The inputs may be mutated; callers should not rely on base or override
// remaining unchanged. Use the returned value.
func deepMerge(base, override any) any {
	baseMap, baseIsMap := base.(map[string]any)
	overrideMap, overrideIsMap := override.(map[string]any)
	if baseIsMap && overrideIsMap {
		for k, ov := range overrideMap {
			if bv, ok := baseMap[k]; ok {
				baseMap[k] = deepMerge(bv, ov)
			} else {
				baseMap[k] = ov
			}
		}
		return baseMap
	}

	baseList, baseIsList := base.([]any)
	overrideList, overrideIsList := override.([]any)
	if baseIsList && overrideIsList {
		return append(baseList, overrideList...)
	}

	return override
}

// applySet sets the value at the given dotted path (e.g. "spec.image") in doc.
// The value string is parsed as YAML, so "3" becomes an int, "true" a bool, and
// "ghcr.io/nais/app:latest" stays a string. Intermediate maps are created as
// needed. It is an error if an intermediate path segment exists but is not a map.
func applySet(doc map[string]any, path, rawValue string) error {
	segments := strings.Split(path, ".")
	if path == "" || len(segments) == 0 {
		return fmt.Errorf("empty --set key")
	}
	if slices.Contains(segments, "") {
		return fmt.Errorf("invalid --set key %q: empty path segment", path)
	}

	var value any
	if err := yaml.Unmarshal([]byte(rawValue), &value); err != nil {
		return fmt.Errorf("invalid --set value for %q: %w", path, err)
	}

	current := doc
	for _, seg := range segments[:len(segments)-1] {
		next, ok := current[seg]
		if !ok {
			child := map[string]any{}
			current[seg] = child
			current = child
			continue
		}
		childMap, ok := next.(map[string]any)
		if !ok {
			return fmt.Errorf("cannot set %q: %q is not a map", path, seg)
		}
		current = childMap
	}

	current[segments[len(segments)-1]] = value
	return nil
}
