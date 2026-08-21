package notification

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var variablePattern = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_.-]+)\s*\}\}`)

func Render(template string, vars map[string]string) (string, error) {
	missing := map[string]struct{}{}
	out := variablePattern.ReplaceAllStringFunc(template, func(v string) string {
		k := strings.TrimSpace(strings.Trim(v, "{}"))
		if val, ok := vars[k]; ok {
			return val
		}
		missing[k] = struct{}{}
		return ""
	})
	if len(missing) > 0 {
		ks := make([]string, 0, len(missing))
		for k := range missing {
			ks = append(ks, k)
		}
		return out, fmt.Errorf("missing template variables: %s", strings.Join(ks, ", "))
	}
	return out, nil
}
func ValidateTemplate(template string) error {
	if strings.Contains(strings.ToLower(template), "<script") {
		return errors.New("unsafe template")
	}
	return nil
}
