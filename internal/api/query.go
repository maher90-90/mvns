package api

import (
	"fmt"
	"strings"
)

func BuildQuery(input string) string {
	input = strings.TrimSpace(input)

	if strings.Contains(input, ":") {
		parts := strings.SplitN(input, ":", 2)
		g := strings.TrimSpace(parts[0])
		a := strings.TrimSpace(parts[1])
		if g != "" && a != "" {
			return fmt.Sprintf(`g:"%s" AND a:"%s"`, g, a)
		}
		if g != "" {
			return fmt.Sprintf(`g:"%s"`, g)
		}
		if a != "" {
			return fmt.Sprintf(`a:"%s"`, a)
		}
	}

	parts := strings.Split(input, ".")
	if len(parts) >= 2 && !strings.Contains(input, " ") {
		allValid := true
		for _, p := range parts {
			if p == "" {
				allValid = false
				break
			}
		}
		if allValid {
			return fmt.Sprintf(`g:"%s"`, input)
		}
	}

	return input
}
