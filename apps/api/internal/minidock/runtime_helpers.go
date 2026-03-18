package minidock

import "strings"

func shortHostnameForWorkload(workloadID string) string {
	value := strings.TrimSpace(workloadID)
	if value == "" {
		return "mdk-ctr"
	}
	value = strings.ReplaceAll(value, "_", "-")
	if len(value) > 24 {
		value = value[:24]
	}
	return value
}
