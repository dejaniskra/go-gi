package http

import (
	"strings"
)

func matchRoute(pattern, actual string) (map[string]string, bool) {
	pParts := strings.Split(strings.Trim(pattern, "/"), "/")
	aParts := strings.Split(strings.Trim(actual, "/"), "/")

	if len(pParts) != len(aParts) {
		return nil, false
	}

	params := make(map[string]string)
	for i := range pParts {
		if strings.HasPrefix(pParts[i], ":") {
			params[strings.TrimPrefix(pParts[i], ":")] = aParts[i]
		} else if pParts[i] != aParts[i] {
			return nil, false
		}
	}

	return params, true
}
