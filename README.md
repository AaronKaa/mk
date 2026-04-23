# mk

`mk` is a small shortcut runner for project-local commands.

Define commands in `mk.json`, `mk.yaml`, or `mk.yml`, then run them globally from inside that project:

```sh
mk
mk test ./...
mk --help
mk edit
```

## Install

```sh
go install github.com/AaronKaa/mk@latest
```

For local development:

```sh
go install .
```

Tagged releases include prebuilt binaries for:
- Linux `amd64`
- Linux `arm64`
- Linux `armv6` for Raspberry Pi Zero / Pi 1
- Linux `armv7` for Raspberry Pi 2/3/4/5 on 32-bit OS
- macOS `amd64` and `arm64`
- Windows `amd64` and `arm64`
- FreeBSD `amd64` and `arm64`

## Config

```json
{
  "header": "Example Project Commands",
  "commands": {
    "test": {
      "command": "go test",
      "open": true,
      "help": "Run tests.",
      "usage": "mk test [packages...]",
      "group": "Quality"
    },
    "build": {
      "command": "go build ./...",
      "open": false,
      "help": "Build the project.",
      "usage": "mk build",
      "group": "Quality"
    },
    "check": {
      "commands": [
        "go test ./...",
        "go vet ./..."
      ],
      "help": "Run all checks.",
      "usage": "mk check",
      "group": "Quality"
    }
  }
}
```

When `open` is `true`, any extra arguments passed to `mk` are shell-quoted and appended to the configured command. For example:

```sh
mk test ./internal/...
```

runs:

```sh
go test ./internal/...
```

If a command needs arguments in the middle of the command, use placeholders:

```json
{
  "command": "echo first={{arg1}} all={{args}}",
  "open": true
}
```

Supported placeholders are `{{args}}`, `{{arg1}}`, `{{arg2}}`, and `{{arg3}}`. When a placeholder is present, arguments are inserted at the placeholder instead of appended.

Use `{{args_prefix "--flag"}}` when a flag should only appear if arguments were provided:

```yaml
commands:
  test:
    commands:
      - go test ./... {{args_prefix "-run"}}
      - go test ./internal/... {{args_prefix "-run"}}
    usage: mk test [filter]
```

`mk test` omits `-run`. `mk test UserTest` expands to `-run UserTest`. Use `{{arg1_prefix "--flag"}}` if only the first argument should be used.

Placeholders also work without `open: true`. In that case, arguments are only inserted where placeholders appear and are not appended to the end:

```json
{
  "command": "cp {{arg1}} dist/{{arg1}}",
  "open": false,
  "help": "Copy one file into dist.",
  "usage": "mk copy <file>"
}
```

Running `mk copy app.js` executes:

```sh
cp app.js dist/app.js
```

Use `commands` instead of `command` for a shortcut that should run multiple commands in order. For multi-command blocks with `open: true`, extra arguments are appended to each command only when arguments are provided:

```json
{
  "commands": [
    "go test",
    "go vet"
  ],
  "open": true
}
```

## More Features

Run dependencies before a command:

```yaml
commands:
  build:
    deps: [generate, lint]
    command: go build ./...
```

Set global or per-command environment variables:

```yaml
env:
  APP_ENV: development
env_file:
  - .env
commands:
  test:
    command: go test ./...
    env_file: .env.test
    env:
      CGO_ENABLED: "0"
```

`env_file` accepts either a string or a list of files. Relative paths are resolved from the directory containing `mk.json` or `mk.yaml`. Files use dotenv-style `KEY=value` lines, with optional `export KEY=value` syntax. Precedence is: global `env_file`, global `env`, command `env_file`, command `env`.

Define reusable variables and expand them with `{{NAME}}`:

```yaml
vars:
  PROGRAM: foo
  CC: cc
  CFLAGS: -Wall -pedantic
commands:
  build:
    command: "{{CC}} {{CFLAGS}} -o {{PROGRAM}}"
```

Variables can also come from shell commands. They are evaluated from the command working directory before the command runs, and are also exposed as environment variables to that command:

```yaml
vars:
  C_FILES:
    shell: "printf '%s' *.c"
commands:
  list-c:
    command: "echo {{C_FILES}}"
```

Command-level variables override global variables:

```yaml
vars:
  TARGET: ./...
commands:
  test-one:
    vars:
      TARGET: ./internal/config
    command: go test {{TARGET}}
```

Run from a specific directory:

```yaml
commands:
  docs:
    dir: website
    command: npm run build
```

Force every command to run from a specific path, or override one command to ignore the global `path_force` with `@`:

```yaml
path_force: tools
commands:
  build-tools:
    command: go build ./...
  local:
    path_force: "@"
    command: pwd
```

`@` only disables the inherited global `path_force` for that command. After that, normal resolution still applies, so the command runs from the config directory unless it also sets `dir`. In YAML, `@` must be quoted as `"@"`.

Customize the help/list title:

```yaml
header: My Project Commands
```

Group commands in the command list:

```yaml
commands:
  test:
    group: Quality
    command: go test ./...
```

Create aliases:

```yaml
commands:
  test:
    command: go test ./...
    aliases: [t]
```

Run multi-command blocks in parallel:

```yaml
commands:
  checks:
    commands:
      - go test ./...
      - go vet ./...
    parallel: true
```

Require confirmation before running a command:

```yaml
commands:
  deploy:
    command: ./scripts/deploy
    confirm: true
```

`mk` searches the current directory and then parent directories for `mk.json`, `mk.yaml`, or `mk.yml`. Commands run from the directory containing the config file.

## Environment Config

You can provide a full command set through the `MK_COMMANDS` environment variable. The value can be JSON or YAML using the same schema as `mk.json`/`mk.yaml`.

```sh
export MK_COMMANDS='{"commands":{"ci-test":{"command":"go test ./...","help":"Run CI tests."}}}'
mk ci-test
```

To generate a shell-safe `MK_COMMANDS` export from the local config:

```sh
mk env
mk env --yaml
```

If no `mk.json`, `mk.yaml`, or `mk.yml` is found, `mk` uses `MK_COMMANDS` directly.

If a config file is found, `mk` merges `MK_COMMANDS` into the file-backed project by default. If any command name or alias conflicts, loading fails with a clear error.

To keep environment commands isolated, set `MK_COMMANDS_PREFIX`; every environment command, dependency, and alias is prefixed to avoid conflicts:

```sh
export MK_COMMANDS='{"commands":{"test":{"command":"npm test","aliases":["t"]}}}'
export MK_COMMANDS_PREFIX='ci:'

mk ci:test
mk ci:t
```

Global `env` and `env_file` from `MK_COMMANDS` apply only to the prefixed environment commands when merged into a file-backed project.

Environment-provided commands are shown in a separate inherited section in `mk` and `mk --help`.

If you want environment commands to remain runnable but stay out of the command list when a local file is present, use `hide` globally or per command:

```yaml
hide: true
commands:
  ci-test:
    command: npm test
```

```yaml
commands:
  ci-test:
    command: npm test
    hide: true
```

`hide` only affects environment commands merged into a file-backed project. Hidden commands still run if you call them directly.

## Commands

```sh
mk
mk init
mk init --yaml
mk edit
mk --convert yaml
mk --convert json
mk --dry-run test ./...
mk completion zsh
mk env
mk convert-make Makefile
mk --help
mk help test
mk test [args...]
```

Running `mk` by itself shows the browsable project command list. Running `mk --help` shows usage plus the same command list.

`mk --convert yaml` converts `mk.json` to `mk.yaml` and removes the original JSON file. `mk --convert json` converts `mk.yaml` or `mk.yml` to `mk.json` and removes the original YAML file. Conversion refuses to overwrite an existing target file.

`mk convert-make [Makefile] [--json|--yaml] [-o path]` converts common Makefiles into mk config. It maps section comments to groups, comments/help output to descriptions, simple Make assignments to `vars`, `$(NAME)` references to `{{NAME}}`, explicit target dependencies to `deps`, `cd dir && ...` to `dir`, strips `@`, and skips the reserved `help` target.

The converter is intentionally best-effort. It warns when it sees Make-only features that need manual review, such as pattern rules, automatic variables like `$@` and `$<`, `include`, conditionals, target-specific variables, and Make functions like `$(wildcard ...)` or `$(patsubst ...)`.

The editor is built with Charmbracelet Bubble Tea and writes back to the existing config format. It opens as a browsable command list:

```text
enter/e edit  n new  d delete  ctrl+s/s save  up/down select  q quit
```

Inside the edit form:

```text
tab next field  shift+tab previous  space toggle open  ctrl+s save  esc command list
```

Aliases are edited as a comma-separated field on the command itself.
