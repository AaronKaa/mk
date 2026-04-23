package app

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/AaronKaa/mk/internal/config"
)

func TestCommandEnvLoadsEnvFilesWithPrecedence(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("A=file\nB=file\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".env.command"), []byte("B=command-file\nC=command-file\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{
		EnvFile: config.EnvFiles{".env"},
		Env:     map[string]string{"A": "global-inline"},
	}
	cmd := config.Command{
		EnvFile: config.EnvFiles{".env.command"},
		Env: map[string]string{
			"C": "command-inline",
			"D": "command-inline",
		},
	}

	got, err := commandEnv(cfg, cmd, dir)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"A": "global-inline",
		"B": "command-file",
		"C": "command-inline",
		"D": "command-inline",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("commandEnv() = %#v, want %#v", got, want)
	}
}
