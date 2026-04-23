package app

import (
	"bytes"
	"strings"
	"testing"

	"github.com/AaronKaa/mk/internal/config"
)

func TestEnvExportPrintsLoadableMKCommands(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	cfg := config.Config{
		Commands: map[string]config.Command{
			"quote": {
				Command: "echo 'hello'",
				Help:    "Print a quoted value.",
			},
		},
	}
	if err := config.Save("mk.json", cfg); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := Run([]string{"env"}, &out, &out); err != nil {
		t.Fatal(err)
	}

	line := strings.TrimSpace(out.String())
	const prefix = "export MK_COMMANDS='"
	if !strings.HasPrefix(line, prefix) || !strings.HasSuffix(line, "'") {
		t.Fatalf("unexpected export line: %s", line)
	}

	value := strings.TrimSuffix(strings.TrimPrefix(line, prefix), "'")
	value = strings.ReplaceAll(value, `'\''`, `'`)
	loaded, err := config.LoadEnvCommands(value)
	if err != nil {
		t.Fatal(err)
	}
	if !config.Equal(cfg, loaded) {
		t.Fatalf("loaded config = %#v, want %#v", loaded, cfg)
	}
}

func TestEnvExportYAML(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := config.Save("mk.json", config.Config{
		Commands: map[string]config.Command{
			"test": {Command: "go test ./..."},
		},
	}); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := Run([]string{"env", "--yaml"}, &out, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "commands:") {
		t.Fatalf("expected yaml payload, got %s", out.String())
	}
}

func TestEnvExportUsesOnlyLocalConfigWhenMKCommandsIsSet(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	if err := config.Save("mk.json", config.Config{
		Commands: map[string]config.Command{
			"local": {Command: "echo local"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	t.Setenv(config.EnvCommandsVar, `{"commands":{"inherited":{"command":"echo inherited"}}}`)

	var out bytes.Buffer
	if err := Run([]string{"env"}, &out, &out); err != nil {
		t.Fatal(err)
	}

	line := strings.TrimSpace(out.String())
	value := strings.TrimSuffix(strings.TrimPrefix(line, "export MK_COMMANDS='"), "'")
	value = strings.ReplaceAll(value, `'\''`, `'`)
	loaded, err := config.LoadEnvCommands(value)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := loaded.Commands["inherited"]; ok {
		t.Fatalf("did not expect inherited env command in exported config: %#v", loaded.Commands)
	}
	if _, ok := loaded.Commands["local"]; !ok {
		t.Fatalf("expected local command in exported config: %#v", loaded.Commands)
	}
}
