package app

import (
	"bytes"
	"strings"
	"testing"

	"github.com/AaronKaa/mk/internal/config"
)

func TestDryRunRunsDepsAndAliasWithoutExecuting(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	cfg := config.Config{
		Commands: map[string]config.Command{
			"setup": {Command: "echo setup", Help: "Setup."},
			"test": {
				Deps:    []string{"setup"},
				Command: "echo test",
				Open:    true,
				Aliases: []string{"t"},
			},
		},
	}
	if err := config.Save("mk.json", cfg); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := Run([]string{"--dry-run", "t", "filter"}, &out, &out); err != nil {
		t.Fatal(err)
	}

	got := strings.TrimSpace(out.String())
	want := "echo setup\necho test filter"
	if got != want {
		t.Fatalf("dry run output = %q, want %q", got, want)
	}
}

func TestDryRunAllowsOptionalPrefixTemplateWithoutOpen(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	cfg := config.Config{
		Commands: map[string]config.Command{
			"test": {
				Command: `go test ./... {{args_prefix "-run"}}`,
				Open:    false,
			},
		},
	}
	if err := config.Save("mk.json", cfg); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := Run([]string{"--dry-run", "test", "UserTest"}, &out, &out); err != nil {
		t.Fatal(err)
	}
	if got, want := strings.TrimSpace(out.String()), "go test ./... -run UserTest"; got != want {
		t.Fatalf("dry run output = %q, want %q", got, want)
	}
}
