package apply

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
