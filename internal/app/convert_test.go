package app

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/AaronKaa/mk/internal/config"
)

func TestConvertConfigRoundTripIsLosslessAndRemovesSource(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	original := config.Config{
		Env:     map[string]string{"APP_ENV": "local"},
		EnvFile: config.EnvFiles{".env"},
		Commands: map[string]config.Command{
			"test": {
				Command: "go test ./...",
				Open:    true,
				Help:    "Run tests.",
				Usage:   "mk test [packages...]",
				Group:   "Quality",
				Env:     map[string]string{"CGO_ENABLED": "0"},
				EnvFile: config.EnvFiles{".env.test"},
				Aliases: []string{"t"},
			},
			"build": {
				Command: "go build ./...",
				Help:    "Build the project.",
				Usage:   "mk build",
				Group:   "Quality",
			},
			"check": {
				Commands: []string{"go test ./...", "go vet ./..."},
				Deps:     []string{"build"},
				Open:     true,
				Help:     "Run checks.",
				Usage:    "mk check [packages...]",
				Group:    "Quality",
				Dir:      "internal",
				Parallel: true,
			},
			"deploy": {
				Command: "./scripts/deploy",
				Help:    "Deploy.",
				Confirm: true,
			},
		},
	}

	if err := config.Save("mk.json", original); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	if err := convertConfig([]string{"yaml"}, &out); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "mk.json")); !os.IsNotExist(err) {
		t.Fatalf("expected mk.json to be removed, stat err: %v", err)
	}
	assertConfigEqualFile(t, original, "mk.yaml")

	out.Reset()
	if err := convertConfig([]string{"json"}, &out); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "mk.yaml")); !os.IsNotExist(err) {
		t.Fatalf("expected mk.yaml to be removed, stat err: %v", err)
	}
	assertConfigEqualFile(t, original, "mk.json")
}

func assertConfigEqualFile(t *testing.T, want config.Config, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
	got, err := config.Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !config.Equal(want, got) {
		t.Fatalf("%s conversion was not lossless: %#v", path, got)
	}
}
