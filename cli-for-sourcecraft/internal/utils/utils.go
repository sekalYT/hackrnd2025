// internal/utils/utils.go
package utils

import (
	"fmt"
	"time"
)

// DerefString safely gets the value of a string pointer, returning "" if nil.
func DerefString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// DerefBool safely gets the value of a bool pointer, returning false if nil.
func DerefBool(b *bool) bool {
	if b != nil {
		return *b
	}
	return false
}

// You can add other common utility functions here later.
func FormatTimeAgo(t time.Time) string {
	duration := time.Since(t)
	if duration.Minutes() < 60 {
		return fmt.Sprintf("%d minutes", int(duration.Minutes()))
	}
	if duration.Hours() < 24 {
		return fmt.Sprintf("%d hours", int(duration.Hours()))
	}
	return fmt.Sprintf("%d days", int(duration.Hours()/24))
}

// FormatRelativeTime (из pr_list.go) - принимает string
func FormatRelativeTime(ts string) string {
	if ts == "" {
		return "-"
	}
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		t, err = time.Parse(time.RFC3339, ts)
		if err != nil {
			return ts
		} // Return raw if both fail
	}

	duration := time.Since(t)
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 7 {
		return t.Local().Format("Jan 02, 2006")
	} else if days > 0 {
		return fmt.Sprintf("%dd ago", days)
	} else if hours > 0 {
		return fmt.Sprintf("%dh ago", hours)
	} else if minutes > 1 {
		return fmt.Sprintf("%dm ago", minutes)
	} else {
		return "Just now"
	}
}
