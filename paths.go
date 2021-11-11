package dispatch

import (
	"fmt"
	"strings"
)

// PathVars is an alias for map[string]string, used for captured path variables.
type PathVars map[string]string

// An APIPath represents a specified path and method, such as GET/users/{uuid}.
type APIPath struct {
	PathParts []string
	Method    string
}

// NewAPIPath creates an APIPath object from a path string, in the format
// GET/users/{uuid}.
func NewAPIPath(path string) (*APIPath, error) {
	parts := strings.Split(path, "/")
	// path must have at least a method and one slash
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid path: %s", path)
	}
	return &APIPath{
		Method:    parts[0],
		PathParts: parts[1:],
	}, nil
}

// Match tests an APIPath against a path string, and returns a map of path
// variables and a boolean representing whether it was a match.
func (a *APIPath) Match(method, path string) (pathVars PathVars, ok bool) {
	if method != a.Method {
		return
	}
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	parts := strings.Split(path, "/")
	if len(parts) != len(a.PathParts) {
		return
	}

	pathVars = make(map[string]string)
	for i, p := range parts {
		apiPart := a.PathParts[i]
		if len(apiPart) > 1 && apiPart[0] == '{' && apiPart[len(apiPart)-1] == '}' {
			// This path part is a path variable
			pathVars[apiPart[1:len(apiPart)-1]] = p
		} else if p != apiPart {
			// If not a path variable, and they don't match, this path is incorrect
			return nil, false
		}
	}
	return pathVars, true
}

// MatchPath tests an APIPath against a path string, and returns true if the
// path matches.
func (a *APIPath) MatchPath(path string) bool {
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	parts := strings.Split(path, "/")
	if len(parts) != len(a.PathParts) {
		return false
	}

	for i, p := range parts {
		apiPart := a.PathParts[i]
		isPathVar := len(apiPart) > 1 && apiPart[0] == '{' && apiPart[len(apiPart)-1] == '}'
		if !isPathVar && p != apiPart {
			return false
		}
	}
	return true
}
