package main

import (
	"path/filepath"
	"testing"
)

func TestResolveDBPathPrefersFlagThenEnv(t *testing.T) {
	t.Setenv("HDTOOLS_DB", "/from/env.db")
	got, err := resolveDBPath("/from/flag.db")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/from/flag.db" {
		t.Fatalf("got %q, want flag path", got)
	}

	got, err = resolveDBPath("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/from/env.db" {
		t.Fatalf("got %q, want env path", got)
	}
}

func TestResolveDBPathDefault(t *testing.T) {
	t.Setenv("HDTOOLS_DB", "")
	got, err := resolveDBPath("")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(got) != "hdtools.db" {
		t.Fatalf("got %q", got)
	}
}
