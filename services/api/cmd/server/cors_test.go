package main

import (
	"reflect"
	"testing"
)

func TestCORSOrigins(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want []string
	}{
		{"empty falls back to local dev", "", []string{"http://localhost:3000"}},
		{"blank falls back to local dev", "  , ,", []string{"http://localhost:3000"}},
		{"single origin", "https://mytax.example.com", []string{"https://mytax.example.com"}},
		{"trims and splits", " https://a.example.com , https://b.example.com ", []string{"https://a.example.com", "https://b.example.com"}},
		{"wildcard is rejected", "*", []string{"http://localhost:3000"}},
		{"wildcard among others is dropped", "https://a.example.com,*", []string{"https://a.example.com"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := corsOrigins(tc.raw); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("corsOrigins(%q) = %v, want %v", tc.raw, got, tc.want)
			}
		})
	}
}
