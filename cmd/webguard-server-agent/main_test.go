package main

import (
	"regexp"
	"testing"
)

func TestVersionHasDevelopmentDefault(t *testing.T) {
	if version != "dev" {
		t.Fatalf("version = %q, want development default %q", version, "dev")
	}
}

func TestNewUUIDReturnsVersionFourUUID(t *testing.T) {
	identifier, err := newUUID()
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(identifier) {
		t.Fatalf("invalid UUID: %q", identifier)
	}
}
