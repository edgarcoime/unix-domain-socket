package pkg

import "strings"

func StringInputParser(s string) string {
	return strings.TrimSpace(strings.Trim(s, "\x00"))
}
