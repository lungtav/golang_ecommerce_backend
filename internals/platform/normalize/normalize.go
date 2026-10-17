package normalize

import "strings"

func Email(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
