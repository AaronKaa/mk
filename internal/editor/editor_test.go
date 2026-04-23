package editor

import (
	"testing"

	"github.com/AaronKaa/mk/internal/config"
)

func selectCommand(m *model, target string) {
	for i, name := range m.names {
		if name == target {
			m.index = i
			return
		}
	}
}

func TestDeleteCurrentRefusesToBreakAlias(t *testing.T) {
	t.Parallel()

	m := newModel("mk.json", config.Config{
		Commands: map[string]config.Command{
			"test": {
				Command: "go test ./...",
			},
			"t": {
				Alias: "test",
			},
		},
	})
	for i, name := range m.names {
		if name == "test" {
			m.index = i
			break
		}
	}

	m.deleteCurrent()
	if m.err == "" {
		t.Fatal("expected validation error")
	}
	if _, ok := m.cfg.Commands["test"]; !ok {
		t.Fatal("expected deleted command to be restored")
	}
}

func TestPersistCurrentStoresAttachedAliases(t *testing.T) {
	t.Parallel()

	m := newModel("mk.json", config.Config{
		Commands: map[string]config.Command{
			"test": {
				Command: "go test ./...",
			},
			"setup": {
				Command: "echo setup",
			},
			"lint": {
				Command: "echo lint",
			},
		},
	})
	selectCommand(&m, "test")
	m.beginEdit()
	m.inputs[2].SetValue("t, test-short")

	m.persistCurrent()
	if m.err != "" {
		t.Fatalf("unexpected error: %s", m.err)
	}

	got := m.cfg.Commands["test"].Aliases
	want := []string{"t", "test-short"}
	if len(got) != len(want) {
		t.Fatalf("aliases = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("aliases = %#v, want %#v", got, want)
		}
	}
}

func TestPersistCurrentStoresCommandProperties(t *testing.T) {
	t.Parallel()

	m := newModel("mk.json", config.Config{
		Commands: map[string]config.Command{
			"test": {
				Command: "go test ./...",
			},
			"setup": {
				Command: "echo setup",
			},
			"lint": {
				Command: "echo lint",
			},
		},
	})
	selectCommand(&m, "test")
	m.beginEdit()
	m.inputs[5].SetValue("Quality")
	m.inputs[6].SetValue("tools")
	m.inputs[7].SetValue("@")
	m.inputs[8].SetValue("setup, lint")
	m.inputs[9].SetValue(".env, .env.local")
	m.inputs[10].SetValue("A=1, B=2")
	m.inputs[11].SetValue("TARGET=./..., FILES:=printf '*.go'")
	m.parallel = true
	m.confirm = true
	m.hide = true

	m.persistCurrent()
	if m.err != "" {
		t.Fatalf("unexpected error: %s", m.err)
	}

	cmd := m.cfg.Commands["test"]
	if cmd.Group != "Quality" || cmd.Dir != "tools" || cmd.PathForce != "@" {
		t.Fatalf("command fields not persisted: %#v", cmd)
	}
	if len(cmd.Deps) != 2 || cmd.Deps[0] != "setup" || cmd.Deps[1] != "lint" {
		t.Fatalf("deps = %#v", cmd.Deps)
	}
	if len(cmd.EnvFile) != 2 || cmd.EnvFile[0] != ".env" || cmd.EnvFile[1] != ".env.local" {
		t.Fatalf("env_file = %#v", cmd.EnvFile)
	}
	if cmd.Env["A"] != "1" || cmd.Env["B"] != "2" {
		t.Fatalf("env = %#v", cmd.Env)
	}
	if cmd.Vars["TARGET"].Value != "./..." || cmd.Vars["FILES"].Shell != "printf '*.go'" {
		t.Fatalf("vars = %#v", cmd.Vars)
	}
	if !cmd.Parallel || !cmd.Confirm || !cmd.Hide {
		t.Fatalf("bool fields not persisted: %#v", cmd)
	}
}

func TestPersistGlobalStoresConfigProperties(t *testing.T) {
	t.Parallel()

	m := newModel("mk.json", config.Config{
		Commands: map[string]config.Command{
			"test": {Command: "go test ./..."},
		},
	})
	m.beginGlobalEdit()
	m.ginputs[0].SetValue("Project Commands")
	m.ginputs[1].SetValue("tools")
	m.ginputs[2].SetValue(".env, .env.local")
	m.ginputs[3].SetValue("APP_ENV=dev, DEBUG=1")
	m.ginputs[4].SetValue("TARGET=./..., FILES:=printf '*.go'")
	m.hide = true

	m.persistGlobal()
	if m.err != "" {
		t.Fatalf("unexpected error: %s", m.err)
	}
	if got := m.cfg.Header; got != "Project Commands" {
		t.Fatalf("Header = %q, want %q", got, "Project Commands")
	}
	if got := m.cfg.PathForce; got != "tools" {
		t.Fatalf("PathForce = %q, want %q", got, "tools")
	}
	if len(m.cfg.EnvFile) != 2 || m.cfg.EnvFile[0] != ".env" || m.cfg.EnvFile[1] != ".env.local" {
		t.Fatalf("EnvFile = %#v", m.cfg.EnvFile)
	}
	if m.cfg.Env["APP_ENV"] != "dev" || m.cfg.Env["DEBUG"] != "1" {
		t.Fatalf("Env = %#v", m.cfg.Env)
	}
	if m.cfg.Vars["TARGET"].Value != "./..." || m.cfg.Vars["FILES"].Shell != "printf '*.go'" {
		t.Fatalf("Vars = %#v", m.cfg.Vars)
	}
	if !m.cfg.Hide {
		t.Fatal("expected Hide to be true")
	}
}
