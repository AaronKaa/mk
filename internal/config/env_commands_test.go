package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFindAndLoadUsesMKCommandsWhenNoFileExists(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(EnvCommandsVar, `{"commands":{"ci-test":{"command":"go test ./...","help":"Run CI tests."}}}`)
	t.Setenv(EnvCommandsPrefixVar, "")

	cfg, path, err := FindAndLoad(dir)
	if err != nil {
		t.Fatal(err)
	}
	if path != EnvCommandsPath {
		t.Fatalf("path = %q, want %q", path, EnvCommandsPath)
	}
	if cfg.Commands["ci-test"].Command != "go test ./..." {
		t.Fatalf("unexpected env command: %#v", cfg.Commands["ci-test"])
	}

	source, err := FindAndLoadSource(dir)
	if err != nil {
		t.Fatal(err)
	}
	if source.BaseDir != dir {
		t.Fatalf("BaseDir = %q, want %q", source.BaseDir, dir)
	}
}

func TestFindAndLoadPrefixedMKCommandsMergeWithFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "mk.json"), []byte(`{"commands":{"test":{"command":"go test ./..."}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(EnvCommandsVar, `{"env":{"NODE_ENV":"test"},"env_file":".env.ci","commands":{"test":{"command":"npm test","aliases":["t"],"env":{"NODE_ENV":"command"}},"build":{"deps":["test"],"command":"npm run build"}}}`)
	t.Setenv(EnvCommandsPrefixVar, "ci:")

	cfg, path, err := FindAndLoad(dir)
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Join(dir, "mk.json") {
		t.Fatalf("path = %q, want file path", path)
	}
	if cfg.Commands["test"].Command != "go test ./..." {
		t.Fatalf("file command was overwritten: %#v", cfg.Commands["test"])
	}
	if cfg.Commands["ci:test"].Command != "npm test" {
		t.Fatalf("missing prefixed env command: %#v", cfg.Commands["ci:test"])
	}
	if cfg.Commands["ci:test"].Env["NODE_ENV"] != "command" {
		t.Fatalf("command env should override env config global env: %#v", cfg.Commands["ci:test"].Env)
	}
	if got := []string(cfg.Commands["ci:test"].EnvFile); !reflect.DeepEqual(got, []string{".env.ci"}) {
		t.Fatalf("env_file was not applied to prefixed command: %#v", got)
	}
	if len(cfg.Env) != 0 {
		t.Fatalf("env config global env leaked into file config: %#v", cfg.Env)
	}
	if got := cfg.Commands["ci:test"].Aliases; !reflect.DeepEqual(got, []string{"ci:t"}) {
		t.Fatalf("alias was not prefixed: %#v", got)
	}
	if got := cfg.Commands["ci:build"].Deps; !reflect.DeepEqual(got, []string{"ci:test"}) {
		t.Fatalf("deps were not prefixed: %#v", got)
	}
	if got := cfg.Commands["ci:test"].Vars; len(got) != 0 {
		t.Fatalf("unexpected vars on ci:test: %#v", got)
	}
}

func TestFindAndLoadMergesMKCommandsWhenFileExistsWithoutPrefix(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "mk.json"), []byte(`{"commands":{"test":{"command":"go test ./..."}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(EnvCommandsVar, `{"commands":{"ci-test":{"command":"npm test"}}}`)
	t.Setenv(EnvCommandsPrefixVar, "")

	cfg, _, err := FindAndLoad(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.Commands["ci-test"].Command; got != "npm test" {
		t.Fatalf("expected env commands to merge when there is no conflict: %#v", cfg.Commands["ci-test"])
	}
}

func TestFindAndLoadSourceMarksInheritedAndHiddenEnvCommands(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "mk.json"), []byte(`{"commands":{"test":{"command":"go test ./..."}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(EnvCommandsVar, `{"hide":true,"commands":{"ci-test":{"command":"npm test"},"visible":{"command":"echo visible","hide":false}}}`)
	t.Setenv(EnvCommandsPrefixVar, "")

	source, err := FindAndLoadSource(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !source.Inherited["ci-test"] || !source.Inherited["visible"] {
		t.Fatalf("expected merged env commands to be marked inherited: %#v", source.Inherited)
	}
	if !source.Hidden["ci-test"] || !source.Hidden["visible"] {
		t.Fatalf("expected global hide to hide inherited commands: %#v", source.Hidden)
	}
}

func TestFindAndLoadSourceAllowsPerCommandHide(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "mk.json"), []byte(`{"commands":{"test":{"command":"go test ./..."}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(EnvCommandsVar, `{"commands":{"ci-test":{"command":"npm test","hide":true},"visible":{"command":"echo visible"}}}`)
	t.Setenv(EnvCommandsPrefixVar, "")

	source, err := FindAndLoadSource(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !source.Hidden["ci-test"] {
		t.Fatalf("expected ci-test to be hidden: %#v", source.Hidden)
	}
	if source.Hidden["visible"] {
		t.Fatalf("did not expect visible to be hidden: %#v", source.Hidden)
	}
}

func TestFindAndLoadMergesEnvGlobalVarsAndPathForceIntoInheritedCommands(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "mk.json"), []byte(`{"commands":{"test":{"command":"go test ./..."}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(EnvCommandsVar, `{"path_force":"tools","vars":{"TARGET":"./..."},"commands":{"build":{"command":"go build {{TARGET}}"}}}`)
	t.Setenv(EnvCommandsPrefixVar, "")

	source, err := FindAndLoadSource(dir)
	if err != nil {
		t.Fatal(err)
	}
	build := source.Config.Commands["build"]
	if got := build.PathForce; got != "tools" {
		t.Fatalf("build path_force = %q, want tools", got)
	}
	if got := build.Vars["TARGET"].Value; got != "./..." {
		t.Fatalf("build TARGET var = %q, want ./...", got)
	}
}
