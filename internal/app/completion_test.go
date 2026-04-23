package app

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/AaronKaa/mk/internal/config"
)

func TestCompletionEscapesNames(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	cfg := config.Config{
		Commands: map[string]config.Command{
			"ci:test": {
				Command: "echo test",
				Help:    "Run 'quoted' test.",
			},
		},
	}
	if err := config.Save("mk.json", cfg); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := Run([]string{"completion", "fish"}, &out, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "'ci:test'") {
		t.Fatalf("fish completion did not quote name: %s", out.String())
	}
	if !strings.Contains(out.String(), `'Run '\''quoted'\'' test.'`) {
		t.Fatalf("fish completion did not escape help: %s", out.String())
	}
}

func TestCompletionOmitsHiddenInheritedCommands(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile("mk.json", []byte(`{"commands":{"local":{"command":"echo local"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(config.EnvCommandsVar, `{"commands":{"visible":{"command":"echo visible"},"secret":{"command":"echo secret","hide":true}}}`)

	var out bytes.Buffer
	if err := Run([]string{"completion", "fish"}, &out, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "'visible'") {
		t.Fatalf("expected visible inherited command in completion: %s", out.String())
	}
	if strings.Contains(out.String(), "'secret'") {
		t.Fatalf("expected hidden inherited command to be omitted: %s", out.String())
	}
}

func TestCompletionReturnsInvalidConfigError(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile("mk.json", []byte(`{"commands":{"bad":{"command":""}}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := Run([]string{"completion", "fish"}, &out, &out); err == nil {
		t.Fatal("expected invalid config error")
	}
}
