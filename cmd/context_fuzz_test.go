package cmd

import (
	"net"
	"testing"
)

func FuzzNormalizeAPIHost(f *testing.F) {
	f.Add("api.infinimesh.local")
	f.Add("api.infinimesh.local:9000")
	f.Add("http://api.infinimesh.local:8080")
	f.Add("[")
	f.Fuzz(func(t *testing.T, host string) {
		got, err := normalizeAPIHost(host)
		if err != nil {
			return
		}
		name, port, splitErr := net.SplitHostPort(got)
		if splitErr != nil || name == "" || port == "" {
			t.Fatalf("normalizeAPIHost(%q)=%q", host, got)
		}
	})
}
