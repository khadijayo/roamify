package validator

import (
	"errors"
	"strings"
)

func RequireFields(fields map[string]string) error {
	var missing []string
	for name, val := range fields {
		if strings.TrimSpace(val) == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return errors.New("missing required fields: " + strings.Join(missing, ", "))
	}
	return nil
}
