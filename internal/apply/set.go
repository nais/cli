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
	return applyFieldOverride(doc, path, rawValue, "set")
}

func applyAppend(doc map[string]any, path, rawValue string) error {
	return applyFieldOverride(doc, path, rawValue, "append")
}

type fieldOverride struct {
	operation string
	value     any
}

func applyFieldOverride(doc map[string]any, path, rawValue, operation string) error {
	segments, err := parseOverridePath(path, operation)
	if err != nil {
		return err
	}

	var value any
	if err := yaml.Unmarshal([]byte(rawValue), &value); err != nil {
		return fmt.Errorf("invalid --%s value for %q: %w", operation, path, err)
	}

	override := fieldOverride{operation: operation, value: value}
	_, err = override.apply(doc, true, segments, path)
	return err
}

func parseOverridePath(path, operation string) ([]setSegment, error) {
	if path == "" {
		return nil, fmt.Errorf("empty --%s key", operation)
	}

	var segments []setSegment
	for part := range strings.SplitSeq(path, ".") {
		if part == "" {
			return nil, fmt.Errorf("invalid --%s key %q: empty path segment", operation, path)
		}
		key, rest, hasIndex := strings.Cut(part, "[")
		if key == "" || strings.Contains(key, "]") {
			return nil, fmt.Errorf("invalid --%s key %q: expected a map key", operation, path)
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
				return nil, fmt.Errorf("invalid --%s key %q: expected a non-negative list index in brackets", operation, path)
			}
			index, err := strconv.Atoi(rawIndex)
			if err != nil {
				return nil, fmt.Errorf("invalid --%s key %q: list index %q is too large", operation, path, rawIndex)
			}
			segments = append(segments, setSegment{index: index, isIndex: true})
			if suffix == "" {
				break
			}
			var next bool
			rest, next = strings.CutPrefix(suffix, "[")
			if !next {
				return nil, fmt.Errorf("invalid --%s key %q: expected '.' or '[' after list index", operation, path)
			}
		}
	}
	return segments, nil
}

func (o fieldOverride) apply(current any, exists bool, segments []setSegment, path string) (any, error) {
	if len(segments) == 0 {
		if o.operation == "append" {
			if !exists {
				current = []any{}
			}
			list, ok := current.([]any)
			if !ok {
				return nil, fmt.Errorf("cannot append %q: target is not a list", path)
			}
			return append(list, o.value), nil
		}
		return o.value, nil
	}

	segment := segments[0]
	if segment.isIndex {
		list, ok := current.([]any)
		if !ok {
			return nil, fmt.Errorf("cannot %s %q: index [%d] requires an existing list", o.operation, path, segment.index)
		}
		if segment.index >= len(list) {
			return nil, fmt.Errorf("cannot %s %q: index [%d] out of range for list of length %d", o.operation, path, segment.index, len(list))
		}
		child, err := o.apply(list[segment.index], true, segments[1:], path)
		if err != nil {
			return nil, err
		}
		list[segment.index] = child
		return list, nil
	}

	mapping, ok := current.(map[string]any)
	if !ok || mapping == nil {
		return nil, fmt.Errorf("cannot %s %q: parent of %q is not a map", o.operation, path, segment.key)
	}
	child, childExists := mapping[segment.key]
	if !childExists && len(segments) > 1 && !segments[1].isIndex {
		child = map[string]any{}
	}
	child, err := o.apply(child, childExists, segments[1:], path)
	if err != nil {
		return nil, err
	}
	mapping[segment.key] = child
	return mapping, nil
}
