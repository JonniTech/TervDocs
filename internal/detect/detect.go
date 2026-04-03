package detect

import "strings"

func IsRouteFile(path string) bool {
	l := strings.ToLower(path)
	return strings.Contains(l, "route") || strings.Contains(l, "router") || strings.Contains(l, "/api/")
}
