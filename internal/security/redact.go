package security

import (
	"regexp"
	"strings"
)

var pemPattern = regexp.MustCompile(`(?s)-----BEGIN [^-]+-----.+?-----END [^-]+-----`)

func Redact(s string) string {
	r := pemPattern.ReplaceAllString(s, "[REDACTED-KEY-MATERIAL]")
	for _, k := range []string{"password", "secret", "token"} {
		r = regexp.MustCompile(`(?i)`+k+`=[^&\s]+`).ReplaceAllString(r, k+"=[REDACTED]")
	}
	return strings.TrimSpace(r)
}
func ConstantTimeEqual(a, b string) bool { return subtleCompare([]byte(a), []byte(b)) }
func subtleCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var d byte
	for i := range a {
		d |= a[i] ^ b[i]
	}
	return d == 0
}
