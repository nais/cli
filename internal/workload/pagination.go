package workload

import "fmt"

// NextCursor checks that a connection with more results can advance.
func NextCursor(hasNext bool, end *string, seen map[string]bool) (*string, error) {
	if !hasNext {
		return nil, nil
	}
	if end == nil || *end == "" || seen[*end] {
		return nil, fmt.Errorf("API reported more results without a new cursor")
	}
	seen[*end] = true
	return end, nil
}
