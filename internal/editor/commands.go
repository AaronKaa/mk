package editor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/AaronKaa/mk/internal/config"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *model) applyFocus() {
	for i := range m.inputs {
		if field(i) == m.focus {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

func (m *model) applyGlobalFocus() {
	for i := range m.ginputs {
		if globalField(i) == m.gfocus {
			m.ginputs[i].Focus()
		} else {
			m.ginputs[i].Blur()
		}
	}
}

func (m *model) blurInputs() {
	for i := range m.inputs {
		m.inputs[i].Blur()
	}
	for i := range m.ginputs {
		m.ginputs[i].Blur()
	}
}

func (m *model) beginEdit() {
	m.screen = screenEdit
	m.focus = fieldName
	m.loadSelected()
	m.err = ""
}

func (m *model) beginGlobalEdit() {
	m.screen = screenGlobal
	m.gfocus = globalFieldHeader
	m.loadGlobal()
	m.err = ""
}

func (m *model) loadSelected() {
	for i := range m.inputs {
		m.inputs[i].SetValue("")
	}
	if len(m.names) == 0 {
		m.applyFocus()
		return
	}
	name := m.names[m.index]
	cmd := m.cfg.Commands[name]
	m.inputs[0].SetValue(name)
	m.inputs[1].SetValue(cmd.DisplayCommand())
	m.inputs[2].SetValue(joinAliases(cmd.Aliases))
	m.inputs[3].SetValue(cmd.Help)
	m.inputs[4].SetValue(cmd.Usage)
	m.inputs[5].SetValue(cmd.Group)
	m.inputs[6].SetValue(cmd.Dir)
	m.inputs[7].SetValue(cmd.PathForce)
	m.inputs[8].SetValue(strings.Join(cmd.Deps, ", "))
	m.inputs[9].SetValue(strings.Join([]string(cmd.EnvFile), ", "))
	m.inputs[10].SetValue(joinStringMap(cmd.Env))
	m.inputs[11].SetValue(joinVars(cmd.Vars))
	m.open = cmd.Open
	m.parallel = cmd.Parallel
	m.confirm = cmd.Confirm
	m.hide = cmd.Hide
	m.applyFocus()
}

func (m *model) loadGlobal() {
	for i := range m.ginputs {
		m.ginputs[i].SetValue("")
	}
	m.ginputs[0].SetValue(m.cfg.Header)
	m.ginputs[1].SetValue(m.cfg.PathForce)
	m.ginputs[2].SetValue(strings.Join([]string(m.cfg.EnvFile), ", "))
	m.ginputs[3].SetValue(joinStringMap(m.cfg.Env))
	m.ginputs[4].SetValue(joinVars(m.cfg.Vars))
	m.hide = m.cfg.Hide
	m.applyGlobalFocus()
}

func (m *model) persistCurrent() {
	if len(m.names) == 0 {
		return
	}
	oldName := m.names[m.index]
	oldCommand := m.cfg.Commands[oldName]
	name := strings.TrimSpace(m.inputs[0].Value())
	command := strings.TrimSpace(m.inputs[1].Value())
	if name == "" || command == "" {
		m.err = "name and command are required"
		return
	}

	delete(m.cfg.Commands, oldName)
	next := oldCommand
	next.Command = command
	next.Commands = nil
	next.Alias = ""
	next.Aliases = splitAliases(m.inputs[2].Value())
	next.Open = m.open
	next.Group = strings.TrimSpace(m.inputs[5].Value())
	next.Dir = strings.TrimSpace(m.inputs[6].Value())
	next.PathForce = strings.TrimSpace(m.inputs[7].Value())
	next.Deps = splitList(m.inputs[8].Value())
	next.EnvFile = config.EnvFiles(splitList(m.inputs[9].Value()))
	next.Env = parseStringMap(m.inputs[10].Value())
	next.Vars = parseVars(m.inputs[11].Value())
	next.Help = strings.TrimSpace(m.inputs[3].Value())
	next.Usage = strings.TrimSpace(m.inputs[4].Value())
	next.Parallel = m.parallel
	next.Confirm = m.confirm
	next.Hide = m.hide
	if oldCommand.IsMulti() && command == oldCommand.DisplayCommand() {
		next.Command = ""
		next.Commands = oldCommand.CommandList()
	}
	m.cfg.Commands[name] = next
	if err := config.Validate(m.cfg); err != nil {
		delete(m.cfg.Commands, name)
		m.cfg.Commands[oldName] = oldCommand
		m.err = err.Error()
		return
	}
	m.names = config.SortedNames(m.cfg)
	for i, candidate := range m.names {
		if candidate == name {
			m.index = i
			break
		}
	}
	m.err = ""
}

func (m *model) persistGlobal() {
	oldHeader := m.cfg.Header
	oldPathForce := m.cfg.PathForce
	oldEnvFile := append(config.EnvFiles(nil), m.cfg.EnvFile...)
	oldEnv := cloneStringMap(m.cfg.Env)
	oldVars := cloneVars(m.cfg.Vars)
	oldHide := m.cfg.Hide

	m.cfg.Header = strings.TrimSpace(m.ginputs[0].Value())
	m.cfg.PathForce = strings.TrimSpace(m.ginputs[1].Value())
	m.cfg.EnvFile = config.EnvFiles(splitList(m.ginputs[2].Value()))
	m.cfg.Env = parseStringMap(m.ginputs[3].Value())
	m.cfg.Vars = parseVars(m.ginputs[4].Value())
	m.cfg.Hide = m.hide

	if err := config.Validate(m.cfg); err != nil {
		m.cfg.Header = oldHeader
		m.cfg.PathForce = oldPathForce
		m.cfg.EnvFile = oldEnvFile
		m.cfg.Env = oldEnv
		m.cfg.Vars = oldVars
		m.cfg.Hide = oldHide
		m.err = err.Error()
		return
	}
	m.err = ""
}

func (m *model) addCommand() {
	name := uniqueName(m.cfg)
	m.cfg.Commands[name] = config.Command{Command: "echo hello", Open: true, Help: "Describe this shortcut.", Usage: "mk " + name + " [args...]"}
	m.names = config.SortedNames(m.cfg)
	for i, candidate := range m.names {
		if candidate == name {
			m.index = i
			break
		}
	}
	m.focus = fieldName
	m.loadSelected()
	m.status = "new command added"
}

func (m *model) deleteCurrent() {
	if len(m.names) == 0 {
		return
	}
	name := m.names[m.index]
	deleted := m.cfg.Commands[name]
	delete(m.cfg.Commands, name)
	if err := config.Validate(m.cfg); err != nil {
		m.cfg.Commands[name] = deleted
		m.err = err.Error()
		return
	}
	m.names = config.SortedNames(m.cfg)
	if m.index >= len(m.names) {
		m.index = len(m.names) - 1
	}
	if len(m.names) == 0 {
		m.addCommand()
		return
	}
	m.loadSelected()
	m.status = "command deleted"
}

func (m model) save() (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenEdit:
		m.persistCurrent()
	case screenGlobal:
		m.persistGlobal()
	}
	if m.err != "" {
		return m, nil
	}
	if err := config.Validate(m.cfg); err != nil {
		m.err = err.Error()
		return m, nil
	}
	if err := config.Save(m.path, m.cfg); err != nil {
		m.err = err.Error()
		return m, nil
	}
	m.status = "saved " + m.path
	m.screen = screenBrowse
	m.blurInputs()
	return m, nil
}

func joinAliases(aliases []string) string {
	return strings.Join(aliases, ", ")
}

func splitAliases(value string) []string {
	return splitList(value)
}

func splitList(value string) []string {
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			items = append(items, item)
		}
	}
	return items
}

func joinStringMap(values map[string]string) string {
	if len(values) == 0 {
		return ""
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+values[key])
	}
	return strings.Join(parts, ", ")
}

func parseStringMap(value string) map[string]string {
	items := splitList(value)
	if len(items) == 0 {
		return nil
	}
	out := map[string]string{}
	for _, item := range items {
		key, val, ok := strings.Cut(item, "=")
		if !ok {
			out[strings.TrimSpace(item)] = ""
			continue
		}
		out[strings.TrimSpace(key)] = strings.TrimSpace(val)
	}
	return out
}

func joinVars(vars config.Vars) string {
	if len(vars) == 0 {
		return ""
	}
	keys := make([]string, 0, len(vars))
	for key := range vars {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		variable := vars[key]
		if variable.Shell != "" {
			parts = append(parts, key+":="+variable.Shell)
			continue
		}
		parts = append(parts, key+"="+variable.Value)
	}
	return strings.Join(parts, ", ")
}

func parseVars(value string) config.Vars {
	items := splitList(value)
	if len(items) == 0 {
		return nil
	}
	out := config.Vars{}
	for _, item := range items {
		if key, val, ok := strings.Cut(item, ":="); ok {
			out[strings.TrimSpace(key)] = config.Variable{Shell: strings.TrimSpace(val)}
			continue
		}
		if key, val, ok := strings.Cut(item, "="); ok {
			out[strings.TrimSpace(key)] = config.Variable{Value: strings.TrimSpace(val)}
			continue
		}
		out[strings.TrimSpace(item)] = config.Variable{}
	}
	return out
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]string, len(values))
	for key, value := range values {
		out[key] = value
	}
	return out
}

func cloneVars(values config.Vars) config.Vars {
	if len(values) == 0 {
		return nil
	}
	out := make(config.Vars, len(values))
	for key, value := range values {
		out[key] = value
	}
	return out
}

func uniqueName(cfg config.Config) string {
	base := "new-command"
	if _, ok := cfg.Commands[base]; !ok {
		return base
	}
	for i := 2; ; i++ {
		name := fmt.Sprintf("%s-%d", base, i)
		if _, ok := cfg.Commands[name]; !ok {
			return name
		}
	}
}
