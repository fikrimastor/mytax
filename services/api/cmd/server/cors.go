package main

import "strings"

const defaultCORSOrigin = "http://localhost:3000"

// corsOrigins parses a comma-separated MYTAX_CORS_ORIGINS value.
// Wildcards are dropped (fail closed); with nothing valid left, only local dev is allowed.
func corsOrigins(raw string) []string {
	var out []string
	for _, o := range strings.Split(raw, ",") {
		o = strings.TrimSpace(o)
		if o == "" || o == "*" {
			continue
		}
		out = append(out, o)
	}
	if len(out) == 0 {
		return []string{defaultCORSOrigin}
	}
	return out
}
