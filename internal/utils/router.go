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

/**********************************************************************************
* Recive il percorso del cliente (e.g. recipes/123/ingredients) e lo converte a un
* percorso constant (recipes/:param/ingredients). Questo e fatto per verificare
* se il cliente e richiedendo un percorso essistente nel programma oppure per
* averiguare quale percorso il cliente sta richiedendo.
*
* Restituisce il percorso constante (recipes/:param/ingredients) e un booleano
* dichiarando se le due stringhe sono le stesse.
***********************************************************************************/
func MatchClientPathToConstant(route string, fixedPath string) (string, bool) {
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
