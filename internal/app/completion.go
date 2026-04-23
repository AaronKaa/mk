package app

import (
	"fmt"
	"io"
	"strings"

	"github.com/AaronKaa/mk/internal/config"
)

func completion(args []string, stdout io.Writer) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: mk completion <bash|zsh|fish>")
	}
	source, err := config.FindAndLoadSource(".")
	if err != nil {
		if !config.IsNotFound(err) {
			return err
		}
		source = config.Source{Config: config.Config{Commands: map[string]config.Command{}}, Inherited: map[string]bool{}, Hidden: map[string]bool{}}
	}
	names := source.VisibleNames()
	switch args[0] {
	case "bash":
		fmt.Fprintln(stdout, `_mk_complete() {`)
		fmt.Fprintf(stdout, `  COMPREPLY=($(compgen -W %s -- "${COMP_WORDS[1]}"))`+"\n", shellSingleQuote(strings.Join(names, " ")))
		fmt.Fprintln(stdout, `}`)
		fmt.Fprintln(stdout, `complete -F _mk_complete mk`)
	case "zsh":
		fmt.Fprintf(stdout, `#compdef mk
_arguments '1:command:(%s)'
`, strings.Join(escapeZshWords(names), " "))
	case "fish":
		for _, name := range names {
			_, cmd, _ := config.ResolveCommand(source.Config, name)
			fmt.Fprintf(stdout, "complete -c mk -f -a %s -d %s\n", shellSingleQuote(name), shellSingleQuote(cmd.Help))
		}
	default:
		return fmt.Errorf("unknown completion shell %q; use bash, zsh, or fish", args[0])
	}
	return nil
}

func shellSingleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func escapeZshWords(words []string) []string {
	escaped := make([]string, len(words))
	for i, word := range words {
		replacer := strings.NewReplacer(`\`, `\\`, `:`, `\:`, ` `, `\ `, `'`, `\'`, `"`, `\"`, `(`, `\(`, `)`, `\)`)
		escaped[i] = replacer.Replace(word)
	}
	return escaped
}
