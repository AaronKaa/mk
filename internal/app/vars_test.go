package app

import (
	"bytes"
	"strings"
	"testing"

	"github.com/AaronKaa/mk/internal/config"
)

func TestDryRunExpandsVars(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	cfg := config.Config{
		Vars: config.Vars{
			"PKG": {Value: "./..."},
		},
		Commands: map[string]config.Command{
			"test": {
				Command: "go test {{PKG}}",
			},
		},
	}
	if err := config.Save("mk.json", cfg); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := Run([]string{"--dry-run", "test"}, &out, &out); err != nil {
		t.Fatal(err)
	}
	if got, want := strings.TrimSpace(out.String()), "go test ./..."; got != want {
		t.Fatalf("dry run output = %q, want %q", got, want)
	}
}

func TestDryRunExpandsShellVars(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	cfg := config.Config{
		Vars: config.Vars{
			"C_FILES": {Shell: "printf 'main.c util.c'"},
		},
		Commands: map[string]config.Command{
			"list": {
				Command: "echo {{C_FILES}}",
			},
		},
	}
	if err := config.Save("mk.json", cfg); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := Run([]string{"--dry-run", "list"}, &out, &out); err != nil {
		t.Fatal(err)
	}
	if got, want := strings.TrimSpace(out.String()), "echo main.c util.c"; got != want {
		t.Fatalf("dry run output = %q, want %q", got, want)
	}
}

func TestCommandVarsOverrideGlobalVars(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	cfg := config.Config{
		Vars: config.Vars{
			"TARGET": {Value: "global"},
		},
		Commands: map[string]config.Command{
			"show": {
				Command: "echo {{TARGET}}",
				Vars: config.Vars{
					"TARGET": {Value: "command"},
				},
			},
		},
	}
	if err := config.Save("mk.json", cfg); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := Run([]string{"--dry-run", "show"}, &out, &out); err != nil {
		t.Fatal(err)
	}
	if got, want := strings.TrimSpace(out.String()), "echo command"; got != want {
		t.Fatalf("dry run output = %q, want %q", got, want)
	}
}
