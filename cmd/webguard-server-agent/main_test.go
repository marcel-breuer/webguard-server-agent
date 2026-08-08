package main

import "testing"

func TestVersionHasDevelopmentDefault(t *testing.T) {
	if version != "dev" {
		t.Fatalf("version = %q, want development default %q", version, "dev")
	}
}
