package pkg

import (
	"regexp"
	"strings"
)

// FormatPhoneToWaMe normalizes Indonesian phone numbers into wa.me format.
func FormatPhoneToWaMe(phone string) string {
	// Remove all non-digit and non-plus characters
	re := regexp.MustCompile(`[^\d\+]+`)
	normalized := re.ReplaceAllString(phone, "")

	switch {
	case strings.HasPrefix(normalized, "+62"):
		return "62" + normalized[3:]
	case strings.HasPrefix(normalized, "0"):
		return "62" + normalized[1:]
	case strings.HasPrefix(normalized, "62"):
		return normalized
	default:
		return ""
	}
}
