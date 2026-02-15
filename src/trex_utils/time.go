package trex_utils

import "time"

// Timestamp returns current time as RFC3339 string
func Timestamp() string {
	return time.Now().Format(time.RFC3339)
}
