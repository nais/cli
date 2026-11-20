package workload

import (
	"fmt"
	"strings"
	"time"
)

// InstanceExit is the last termination observed by the API, independent of the current state.
type InstanceExit struct {
	Reason    *string    `json:"reason,omitempty"`
	Code      *int       `json:"code,omitempty"`
	Timestamp *time.Time `json:"timestamp,omitempty"`
}

func LastExit(reason *string, code *int, timestamp *time.Time) *InstanceExit {
	if reason == nil && code == nil && timestamp == nil {
		return nil
	}
	return &InstanceExit{Reason: reason, Code: code, Timestamp: timestamp}
}

func (e InstanceExit) String() string {
	var parts []string
	if e.Reason != nil && *e.Reason != "" {
		parts = append(parts, *e.Reason)
	}
	if e.Code != nil {
		parts = append(parts, fmt.Sprintf("code %d", *e.Code))
	}
	if e.Timestamp != nil {
		parts = append(parts, e.Timestamp.Format(time.RFC3339))
	}
	return strings.Join(parts, ", ")
}
