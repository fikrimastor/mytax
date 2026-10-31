package main

import "strings"

const defaultCORSOrigin = "http://localhost:3000"

// corsOrigins parses a comma-separated MYTAX_CORS_ORIGINS value.
// Any entry containing "*" is dropped (Echo expands patterns like https://*.example.com),
// trailing slashes are trimmed (browsers send Origin without one), and with nothing
// valid left only local dev is allowed.
func corsOrigins(raw string) []string {
	var out []string
	for _, o := range strings.Split(raw, ",") {
		o = strings.TrimRight(strings.TrimSpace(o), "/")
		if o == "" || strings.Contains(o, "*") {
			continue
		}
		out = append(out, o)
	}
	if len(out) == 0 {
		return []string{defaultCORSOrigin}
	}
	return out
}
