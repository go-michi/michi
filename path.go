package michi

import (
	"strings"
)

func methodAndPath(pattern string) (string, string) {
	// https://github.com/golang/go/blob/48b10c9af7955bcab179b60a148a633a0a75cde7/src/net/http/pattern.go#L95-L102
	method, rest, found := pattern, "", false
	if i := strings.IndexAny(pattern, " \t"); i >= 0 {
		method, rest, found = pattern[:i], strings.TrimLeft(pattern[i+1:], " \t"), true
	}
	if !found {
		rest = method
		method = ""
	}
	return method, rest
}

func joinMethodAndPath(method, path string) string {
	if method == "" {
		return path
	}
	return method + " " + path
}

func joinPathAndPattern(path, pattern string) string {
	fullPath := path + pattern
	return strings.ReplaceAll(fullPath, "//", "/")
}
