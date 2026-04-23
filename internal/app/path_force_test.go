package app

import (
	"path/filepath"
	"testing"

	"github.com/AaronKaa/mk/internal/config"
)

func TestCommandDirUsesPathForce(t *testing.T) {
	source := config.Source{
		Config:  config.Config{PathForce: "tools"},
		BaseDir: "/repo",
	}
	if got, want := commandDir(source, config.Command{}), filepath.Join("/repo", "tools"); got != want {
		t.Fatalf("commandDir() = %q, want %q", got, want)
	}
}

func TestCommandDirPathForceAtIgnoresGlobalPathForce(t *testing.T) {
	source := config.Source{
		Config:  config.Config{PathForce: "tools"},
		BaseDir: "/repo",
	}
	if got, want := commandDir(source, config.Command{PathForce: "@"}), "/repo"; got != want {
		t.Fatalf("commandDir() = %q, want %q", got, want)
	}
}

func TestCommandDirPathForceAtStillAllowsCommandDir(t *testing.T) {
	source := config.Source{
		Config:  config.Config{PathForce: "tools"},
		BaseDir: "/repo",
	}
	if got, want := commandDir(source, config.Command{PathForce: "@", Dir: "pkg/api"}), filepath.Join("/repo", "pkg/api"); got != want {
		t.Fatalf("commandDir() = %q, want %q", got, want)
	}
}
