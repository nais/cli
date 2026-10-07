package apply

import (
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type setSegment struct {
	key     string
	index   int
	isIndex bool
}

// applySet replaces a YAML value at a dotted path with optional list indices.
// Missing maps are created, but indexed lists and their elements must exist.
func applySet(doc map[string]any, path, rawValue string) error {
	segments, err := parseSetPath(path)
	if err != nil {
		return err
	}

	var value any
	if err := yaml.Unmarshal([]byte(rawValue), &value); err != nil {
		return fmt.Errorf("invalid --set value for %q: %w", path, err)
	}

	_, err = setPathValue(doc, segments, value, path)
	return err
}

func parseSetPath(path string) ([]setSegment, error) {
	if path == "" {
		return nil, fmt.Errorf("empty --set key")
	}

	var segments []setSegment
	for _, part := range strings.Split(path, ".") {
		if part == "" {
			return nil, fmt.Errorf("invalid --set key %q: empty path segment", path)
		}
		key, rest, hasIndex := strings.Cut(part, "[")
		if key == "" || strings.Contains(key, "]") {
			return nil, fmt.Errorf("invalid --set key %q: expected a map key", path)
		}
		segments = append(segments, setSegment{key: key})
		if !hasIndex {
			continue
		}

		for {
			rawIndex, suffix, closed := strings.Cut(rest, "]")
			if !closed || rawIndex == "" || strings.IndexFunc(rawIndex, func(r rune) bool {
				return r < '0' || r > '9'
			}) != -1 {
				return nil, fmt.Errorf("invalid --set key %q: expected a non-negative list index in brackets", path)
			}
			index, err := strconv.Atoi(rawIndex)
			if err != nil {
				return nil, fmt.Errorf("invalid --set key %q: list index %q is too large", path, rawIndex)
			}
			segments = append(segments, setSegment{index: index, isIndex: true})
			if suffix == "" {
				break
			}
			var next bool
			rest, next = strings.CutPrefix(suffix, "[")
			if !next {
				return nil, fmt.Errorf("invalid --set key %q: expected '.' or '[' after list index", path)
			}
		}
	}
	return segments, nil
}

func setPathValue(current any, segments []setSegment, value any, path string) (any, error) {
	if len(segments) == 0 {
		return value, nil
	}

	segment := segments[0]
	if segment.isIndex {
		list, ok := current.([]any)
		if !ok {
			return nil, fmt.Errorf("cannot set %q: index [%d] requires an existing list", path, segment.index)
		}
		if segment.index >= len(list) {
			return nil, fmt.Errorf("cannot set %q: index [%d] out of range for list of length %d", path, segment.index, len(list))
		}
		child, err := setPathValue(list[segment.index], segments[1:], value, path)
		if err != nil {
			return nil, err
		}
		list[segment.index] = child
		return list, nil
	}

	mapping, ok := current.(map[string]any)
	if !ok || mapping == nil {
		return nil, fmt.Errorf("cannot set %q: parent of %q is not a map", path, segment.key)
	}
	child, exists := mapping[segment.key]
	if !exists && len(segments) > 1 && !segments[1].isIndex {
		child = map[string]any{}
	}
	child, err := setPathValue(child, segments[1:], value, path)
	if err != nil {
		return nil, err
	}
	mapping[segment.key] = child
	return mapping, nil
}
