package utils

import "strings"

// takes in a route path with a dynamic value and maps it to a fixed route string
func MakeRouteFromPath(path string, replaceWith string) string {
	// split each part of the string
	splitPath := strings.Split(path, "/")
	fixedPath := make([]string, 0)

	for _, pathPart := range splitPath {
		hasPrefix := strings.HasPrefix(path, ":")

		if !hasPrefix {
			fixedPath = append(fixedPath, pathPart)
		} else {
			fixedPath = append(fixedPath, replaceWith)
		}

	}

	return strings.Join(fixedPath, "/")
}

// Receives a dynamic rout and maps it to a fixed path (e.g. recipes/123/ingredients -> recipes/:param/ingredients)
// it returns the newly formed path and whether the two are the same.
func MakePathFromRoute(route string, fixedPath string) (string, bool) {
	routeParts := strings.Split(route, "/")
	fixedPathParts := strings.Split(fixedPath, "/")

	mappedPathParts := make([]string, 0)

	for i, part := range routeParts {

		// if the index is out of range these two paths are not the same, return the original route
		if i >= len(fixedPathParts) {
			return route, false
		}

		if part == fixedPathParts[i] {
			mappedPathParts = append(mappedPathParts, part)
		} else {
			mappedPathParts = append(mappedPathParts, fixedPathParts[i])
		}
	}

	mappedPath := strings.Join(mappedPathParts, "/")
	return mappedPath, mappedPath == fixedPath
}
