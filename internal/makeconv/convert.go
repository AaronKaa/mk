package makeconv

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/AaronKaa/mk/internal/config"
)

var (
	targetPattern             = regexp.MustCompile(`^([A-Za-z0-9_.-]+)\s*:(.*)$`)
	assignmentPattern         = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)\s*[:?+]?=\s*(.*)$`)
	targetAssignmentPattern   = regexp.MustCompile(`^[^:]+:\s*[A-Za-z_][A-Za-z0-9_]*\s*=`)
	sectionPattern            = regexp.MustCompile(`^#\s*[─\-=]*\s*(.+?)\s*[─\-=]*\s*$`)
	helpEchoPattern           = regexp.MustCompile(`make\s+([A-Za-z0-9_.-]+)\s+(.+)$`)
	makeVariablePattern       = regexp.MustCompile(`\$\(([A-Za-z_][A-Za-z0-9_]*)\)`)
	makeFunctionPattern       = regexp.MustCompile(`\$\((wildcard|patsubst|foreach|shell|subst|filter|filter-out|addprefix|addsuffix|notdir|dir|basename|suffix|sort|word|wordlist|words|firstword|lastword|if|or|and)\b`)
	automaticVariablePattern  = regexp.MustCompile(`\$[@<*?^+]`)
	conditionalPattern        = regexp.MustCompile(`^(ifn?eq|ifn?def|else|endif)\b`)
	includePattern            = regexp.MustCompile(`^-?include\b`)
	unsupportedSpecialTargets = map[string]bool{
		".PHONY":           true,
		".SUFFIXES":        true,
		".DEFAULT":         true,
		".PRECIOUS":        true,
		".INTERMEDIATE":    true,
		".SECONDARY":       true,
		".DELETE_ON_ERROR": true,
	}
)

type target struct {
	name     string
	group    string
	comments []string
	deps     []string
	recipes  []string
}

func Convert(r io.Reader) (config.Config, []string, error) {
	scanner := bufio.NewScanner(r)
	var (
		targets         []target
		current         *target
		group           string
		vars            = config.Vars{}
		pendingComments []string
		helpLines       []string
		header          string
		warnings        []string
		skipRecipes     bool
		lineNumber      int
	)

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "#") {
			text := strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
			if header == "" && text != "" && !looksLikeSection(text) {
				header = text
			}
			if section := parseSection(trimmed); section != "" {
				group = section
				pendingComments = nil
				continue
			}
			if text != "" {
				pendingComments = append(pendingComments, text)
			}
			continue
		}

		if strings.TrimSpace(line) == "" {
			pendingComments = nil
			current = nil
			skipRecipes = false
			continue
		}

		if conditionalPattern.MatchString(trimmed) {
			warnings = append(warnings, fmt.Sprintf("line %d: skipped GNU Make conditional %q", lineNumber, trimmed))
			current = nil
			pendingComments = nil
			continue
		}

		if includePattern.MatchString(trimmed) {
			warnings = append(warnings, fmt.Sprintf("line %d: skipped include directive %q", lineNumber, trimmed))
			current = nil
			pendingComments = nil
			continue
		}

		if targetAssignmentPattern.MatchString(trimmed) {
			warnings = append(warnings, fmt.Sprintf("line %d: skipped target-specific variable %q", lineNumber, trimmed))
			current = nil
			pendingComments = nil
			continue
		}

		if assignment := assignmentPattern.FindStringSubmatch(trimmed); assignment != nil {
			value := strings.TrimSpace(assignment[2])
			vars[assignment[1]] = config.Variable{Value: rewriteMakeVariables(value)}
			if makeFunctionPattern.MatchString(value) {
				warnings = append(warnings, fmt.Sprintf("line %d: variable %q contains a Make function and may need manual conversion", lineNumber, assignment[1]))
			}
			current = nil
			pendingComments = nil
			continue
		}

		if isRecipeLine(line) {
			if current == nil {
				if skipRecipes {
					continue
				}
				return config.Config{}, nil, fmt.Errorf("line %d: recipe without target", lineNumber)
			}
			recipe := normalizeRecipe(line)
			if automaticVariablePattern.MatchString(recipe) {
				warnings = append(warnings, fmt.Sprintf("line %d: recipe for %q uses Make automatic variables and may need manual conversion", lineNumber, current.name))
			}
			if makeFunctionPattern.MatchString(recipe) {
				warnings = append(warnings, fmt.Sprintf("line %d: recipe for %q uses a Make function and may need manual conversion", lineNumber, current.name))
			}
			if current.name == "help" {
				helpLines = append(helpLines, recipe)
			} else if recipe != "" {
				current.recipes = append(current.recipes, rewriteMakeVariables(recipe))
			}
			continue
		}

		if strings.Contains(trimmed, "%") && strings.Contains(trimmed, ":") {
			warnings = append(warnings, fmt.Sprintf("line %d: skipped pattern rule %q", lineNumber, trimmed))
			current = nil
			skipRecipes = true
			pendingComments = nil
			continue
		}

		if strings.Contains(trimmed, "$(") && strings.Contains(trimmed, ":") {
			warnings = append(warnings, fmt.Sprintf("line %d: skipped dynamic target %q", lineNumber, trimmed))
			current = nil
			skipRecipes = true
			pendingComments = nil
			continue
		}

		match := targetPattern.FindStringSubmatch(trimmed)
		if match == nil {
			current = nil
			pendingComments = nil
			continue
		}
		name := match[1]
		if unsupportedSpecialTargets[name] {
			current = nil
			skipRecipes = false
			pendingComments = nil
			continue
		}
		targets = append(targets, target{name: name, group: group, comments: append([]string(nil), pendingComments...), deps: parseDeps(match[2])})
		current = &targets[len(targets)-1]
		skipRecipes = false
		pendingComments = nil
	}
	if err := scanner.Err(); err != nil {
		return config.Config{}, nil, err
	}

	helpByTarget := parseHelpDescriptions(helpLines)
	cfg := config.Config{Header: header, Vars: vars, Commands: map[string]config.Command{}}
	convertedTargets := map[string]bool{}
	for _, target := range targets {
		if target.name != "help" && (len(target.recipes) > 0 || len(target.deps) > 0) {
			convertedTargets[target.name] = true
		}
	}
	for _, target := range targets {
		if target.name == "help" {
			warnings = append(warnings, `skipped reserved target "help"; mk provides help automatically`)
			continue
		}
		if len(target.recipes) == 0 && len(target.deps) == 0 {
			continue
		}
		cmd := commandFromTarget(target, convertedTargets)
		if help := helpByTarget[target.name]; help != "" {
			cmd.Help = help
		}
		if len(target.recipes) == 0 && len(cmd.Deps) == 0 {
			continue
		}
		if len(cmd.CommandList()) == 1 && strings.TrimSpace(cmd.Command) == "" && len(cmd.Deps) > 0 {
			cmd.Command = ":"
		}
		cfg.Commands[target.name] = cmd
	}
	return cfg, warnings, nil
}

func commandFromTarget(t target, convertedTargets map[string]bool) config.Command {
	cmd := config.Command{
		Help:  strings.Join(t.comments, " "),
		Group: t.group,
	}
	for _, dep := range t.deps {
		if convertedTargets[dep] {
			cmd.Deps = append(cmd.Deps, dep)
		}
	}
	recipes := append([]string(nil), t.recipes...)
	dir, stripped, ok := commonCD(recipes)
	if ok {
		cmd.Dir = dir
		recipes = stripped
	}
	if len(recipes) == 1 {
		cmd.Command = recipes[0]
	} else {
		cmd.Commands = recipes
	}
	return cmd
}

func parseDeps(value string) []string {
	fields := strings.Fields(value)
	deps := make([]string, 0, len(fields))
	for _, field := range fields {
		if strings.Contains(field, "=") || strings.HasPrefix(field, "#") {
			break
		}
		deps = append(deps, field)
	}
	return deps
}

func rewriteMakeVariables(value string) string {
	return makeVariablePattern.ReplaceAllString(value, "{{$1}}")
}

func isRecipeLine(line string) bool {
	return strings.HasPrefix(line, "\t") || strings.HasPrefix(line, "        ")
}

func normalizeRecipe(line string) string {
	recipe := strings.TrimSpace(line)
	recipe = strings.TrimPrefix(recipe, "@")
	if strings.HasPrefix(recipe, "-") {
		recipe = strings.TrimSpace(strings.TrimPrefix(recipe, "-")) + " || true"
	}
	return recipe
}

func parseSection(line string) string {
	if !strings.Contains(line, "─") && !strings.Contains(line, "---") && !strings.Contains(line, "===") {
		return ""
	}
	match := sectionPattern.FindStringSubmatch(line)
	if match == nil {
		return ""
	}
	text := strings.TrimSpace(match[1])
	if text == "" || strings.Contains(text, "===") {
		return ""
	}
	return text
}

func looksLikeSection(text string) bool {
	return strings.ContainsAny(text, "─=")
}

func parseHelpDescriptions(lines []string) map[string]string {
	out := map[string]string{}
	for _, line := range lines {
		line = strings.TrimPrefix(line, "echo ")
		line = strings.Trim(line, `"`)
		match := helpEchoPattern.FindStringSubmatch(line)
		if match != nil {
			out[match[1]] = strings.TrimSpace(match[2])
		}
	}
	return out
}

func commonCD(recipes []string) (string, []string, bool) {
	if len(recipes) == 0 {
		return "", nil, false
	}
	var dir string
	stripped := make([]string, len(recipes))
	for i, recipe := range recipes {
		left, right, ok := strings.Cut(recipe, " && ")
		if !ok || !strings.HasPrefix(left, "cd ") {
			return "", nil, false
		}
		currentDir := strings.TrimSpace(strings.TrimPrefix(left, "cd "))
		if currentDir == "" {
			return "", nil, false
		}
		if dir == "" {
			dir = currentDir
		}
		if dir != currentDir {
			return "", nil, false
		}
		stripped[i] = right
	}
	return dir, stripped, true
}
