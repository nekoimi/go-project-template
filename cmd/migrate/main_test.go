package main

import (
	"io"
	"strings"
	"testing"

	"github.com/nekoimi/go-project-template/internal/config"
)

func TestDatabaseURL(t *testing.T) {
	got := databaseURL(config.DatabaseConfig{
		Host: "2001:db8::1", Port: "5432", User: "solo@pay", Password: "p:a/s", DBName: "solo_pay", SSLMode: "require",
	})
	want := "postgres://solo%40pay:p%3Aa%2Fs@[2001:db8::1]:5432/solo_pay?sslmode=require"
	if got != want {
		t.Fatalf("databaseURL() = %q, want %q", got, want)
	}
}

func TestParseArgs(t *testing.T) {
	opts, command, args, err := parseArgs([]string{"--config", "config/config.test.yaml", "down", "2"}, io.Discard)
	if err != nil {
		t.Fatalf("parseArgs: %v", err)
	}
	if opts.configPath != "config/config.test.yaml" || command != "down" || len(args) != 1 || args[0] != "2" {
		t.Fatalf("unexpected parse result: opts=%+v command=%q args=%v", opts, command, args)
	}
}

func TestPositiveIntArg(t *testing.T) {
	if _, err := positiveIntArg("down", []string{"0"}); err == nil || !strings.Contains(err.Error(), "greater than zero") {
		t.Fatalf("expected positive-step validation error, got %v", err)
	}
}
