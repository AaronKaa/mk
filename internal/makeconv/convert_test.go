package makeconv

import (
	"strings"
	"testing"
)

func TestConvertMakefile(t *testing.T) {
	input := `# Project Header

# ─── Installation ────────────────────────────────────────────
# Guided interactive installer
install:
	@./install.sh

# ─── Build ───────────────────────────────────────────────────
build-rust:
	cd rust-port/wifi-densepose-rs && cargo build --release

clean:
	rm -f .install.log
	cd rust-port/wifi-densepose-rs && cargo clean 2>/dev/null || true

help:
	@echo "  Installation:"
	@echo "    make install          Interactive guided installer"
	@echo "    make build-rust       Build Rust workspace (release)"
`

	cfg, warnings, err := Convert(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Header != "Project Header" {
		t.Fatalf("Header = %q", cfg.Header)
	}
	if cfg.Commands["install"].Group != "Installation" {
		t.Fatalf("install group = %q", cfg.Commands["install"].Group)
	}
	if cfg.Commands["install"].Help != "Interactive guided installer" {
		t.Fatalf("install help = %q", cfg.Commands["install"].Help)
	}
	if cfg.Commands["install"].Command != "./install.sh" {
		t.Fatalf("install command = %q", cfg.Commands["install"].Command)
	}
	if cfg.Commands["build-rust"].Dir != "rust-port/wifi-densepose-rs" {
		t.Fatalf("build-rust dir = %q", cfg.Commands["build-rust"].Dir)
	}
	if cfg.Commands["build-rust"].Command != "cargo build --release" {
		t.Fatalf("build-rust command = %q", cfg.Commands["build-rust"].Command)
	}
	if _, ok := cfg.Commands["help"]; ok {
		t.Fatal("help target should be skipped")
	}
	if len(warnings) != 1 {
		t.Fatalf("warnings = %#v", warnings)
	}
}

func TestConvertMakefileVarsDepsAndWarnings(t *testing.T) {
	input := `PROGRAM = foo
CC = cc
C_FILES := $(wildcard *.c)
OBJS := $(patsubst %.c, %.o, $(C_FILES))

all: build

build: clean
	$(CC) -o $(PROGRAM) $(OBJS)

clean:
	rm -f $(OBJS)

%.o: %.c
	$(CC) -c $< -o $@
`

	cfg, warnings, err := Convert(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.Vars["PROGRAM"].Value; got != "foo" {
		t.Fatalf("PROGRAM = %q, want foo", got)
	}
	if got := cfg.Commands["build"].Command; got != "{{CC}} -o {{PROGRAM}} {{OBJS}}" {
		t.Fatalf("build command = %q", got)
	}
	if got := cfg.Commands["build"].Deps; len(got) != 1 || got[0] != "clean" {
		t.Fatalf("build deps = %#v, want [clean]", got)
	}
	if got := cfg.Commands["clean"].Command; got != "rm -f {{OBJS}}" {
		t.Fatalf("clean command = %q", got)
	}
	if len(warnings) == 0 {
		t.Fatal("expected warnings for Make-specific constructs")
	}
}

func TestConvertGenericGNUMakefileBestEffort(t *testing.T) {
	input := `PROGRAM = foo
C_FILES := $(wildcard *.c)
OBJS := $(patsubst %.c, %.o, $(C_FILES))
CC = cc
CFLAGS = -Wall -pedantic

all: $(PROGRAM)

$(PROGRAM): .depend $(OBJS)
	$(CC) $(CFLAGS) $(OBJS) -o $(PROGRAM)

depend: .depend

.depend: cmd = gcc -MM -MF depend $(var); cat depend >> .depend;
.depend:
	@echo "Generating dependencies..."
	@$(foreach var, $(C_FILES), $(cmd))
	@rm -f depend

-include .depend

%.o: %.c
	$(CC) $(CFLAGS) -c $< -o $@

clean:
	rm -f .depend $(OBJS)

.PHONY: clean depend
`

	cfg, warnings, err := Convert(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.Vars["PROGRAM"].Value; got != "foo" {
		t.Fatalf("PROGRAM = %q, want foo", got)
	}
	if _, ok := cfg.Commands["all"]; ok {
		t.Fatal("all should be skipped because its only dependency is dynamic")
	}
	if got := cfg.Commands["depend"].Deps; len(got) != 1 || got[0] != ".depend" {
		t.Fatalf("depend deps = %#v, want [.depend]", got)
	}
	if got := cfg.Commands[".depend"].Commands; len(got) != 3 {
		t.Fatalf(".depend commands = %#v, want 3 commands", got)
	}
	if got := cfg.Commands["clean"].Command; got != "rm -f .depend {{OBJS}}" {
		t.Fatalf("clean command = %q", got)
	}
	if len(warnings) < 5 {
		t.Fatalf("warnings = %#v, want warnings for skipped Make-specific features", warnings)
	}
}
