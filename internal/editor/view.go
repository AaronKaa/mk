package editor

import (
	"fmt"
	"strings"
)

func (m model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder
	b.WriteString(hotStyle.Render("mk editor"))
	b.WriteString(" ")
	b.WriteString(dimStyle.Render(m.path))
	b.WriteString("\n\n")
	if m.screen == screenEdit {
		b.WriteString(m.editView())
	} else if m.screen == screenGlobal {
		b.WriteString(m.globalView())
	} else {
		b.WriteString(m.browseView())
	}
	if m.status != "" {
		b.WriteString("\n")
		b.WriteString(hotStyle.Render(m.status))
	}
	if m.err != "" {
		b.WriteString("\n")
		b.WriteString(errStyle.Render(m.err))
	}
	return panelStyle.Render(b.String())
}

func (m model) browseView() string {
	var b strings.Builder
	b.WriteString(dimStyle.Render("Project"))
	b.WriteString("\n")
	b.WriteString(m.projectSummary())
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("Commands"))
	b.WriteString("\n")
	b.WriteString(m.commandList())
	b.WriteString("\n\n")
	if len(m.names) > 0 {
		name := m.names[m.index]
		cmd := m.cfg.Commands[name]
		b.WriteString(hotStyle.Render(name))
		b.WriteString("\n")
		if cmd.Help != "" {
			b.WriteString(cmd.Help)
			b.WriteString("\n")
		}
		if cmd.Usage != "" {
			b.WriteString(dimStyle.Render("usage: "))
			b.WriteString(cmd.Usage)
			b.WriteString("\n")
		}
		if len(cmd.Aliases) > 0 {
			b.WriteString(dimStyle.Render("aliases: "))
			b.WriteString(strings.Join(cmd.Aliases, ", "))
			b.WriteString("\n")
		}
		if cmd.Group != "" {
			b.WriteString(dimStyle.Render("group: "))
			b.WriteString(cmd.Group)
			b.WriteString("\n")
		}
		if cmd.Dir != "" {
			b.WriteString(dimStyle.Render("dir: "))
			b.WriteString(cmd.Dir)
			b.WriteString("\n")
		}
		if cmd.PathForce != "" {
			b.WriteString(dimStyle.Render("path_force: "))
			b.WriteString(cmd.PathForce)
			b.WriteString("\n")
		}
		b.WriteString(dimStyle.Render("command: "))
		b.WriteString(cmd.DisplayCommand())
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("open: "))
		b.WriteString(fmt.Sprintf("%t", cmd.Open))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("parallel: "))
		b.WriteString(fmt.Sprintf("%t", cmd.Parallel))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("confirm: "))
		b.WriteString(fmt.Sprintf("%t", cmd.Confirm))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("hide: "))
		b.WriteString(fmt.Sprintf("%t", cmd.Hide))
		b.WriteString("\n")
	} else {
		b.WriteString(dimStyle.Render("No commands yet. Press n to create one."))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("g globals  enter/e edit  n new  d delete  ctrl+s/s save  up/down select  q quit"))
	return b.String()
}

func (m model) editView() string {
	var b strings.Builder
	if len(m.names) > 0 {
		b.WriteString(dimStyle.Render("Editing "))
		b.WriteString(hotStyle.Render(m.names[m.index]))
		b.WriteString("\n\n")
	}
	b.WriteString(m.form())
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("tab next field  shift+tab previous  space toggle bool  ctrl+s save  esc command list"))
	return b.String()
}

func (m model) globalView() string {
	var b strings.Builder
	b.WriteString(dimStyle.Render("Editing "))
	b.WriteString(hotStyle.Render("global properties"))
	b.WriteString("\n\n")
	b.WriteString(m.globalForm())
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("tab next field  shift+tab previous  space toggle bool  ctrl+s save  esc command list"))
	return b.String()
}

func (m model) commandList() string {
	if len(m.names) == 0 {
		return dimStyle.Render("No commands.")
	}
	lines := make([]string, len(m.names))
	for i, name := range m.names {
		prefix := "  "
		style := dimStyle
		if i == m.index {
			prefix = "> "
			style = hotStyle
		}
		lines[i] = prefix + style.Render(name)
	}
	return strings.Join(lines, "\n")
}

func (m model) form() string {
	rows := []string{
		m.label(fieldName, "name") + m.inputs[0].View(),
		m.label(fieldCommand, "command") + m.inputs[1].View(),
		m.label(fieldAliases, "aliases") + m.inputs[2].View(),
		m.label(fieldHelp, "help") + m.inputs[3].View(),
		m.label(fieldUsage, "usage") + m.inputs[4].View(),
		m.label(fieldGroup, "group") + m.inputs[5].View(),
		m.label(fieldDir, "dir") + m.inputs[6].View(),
		m.label(fieldPathForce, "path_force") + m.inputs[7].View(),
		m.label(fieldDeps, "deps") + m.inputs[8].View(),
		m.label(fieldEnvFile, "env_file") + m.inputs[9].View(),
		m.label(fieldEnv, "env") + m.inputs[10].View(),
		m.label(fieldVars, "vars") + m.inputs[11].View(),
		m.label(fieldOpen, "open") + checkbox(m.open),
		m.label(fieldParallel, "parallel") + checkbox(m.parallel),
		m.label(fieldConfirm, "confirm") + checkbox(m.confirm),
		m.label(fieldHideCommand, "hide") + checkbox(m.hide),
	}
	return strings.Join(rows, "\n")
}

func (m model) globalForm() string {
	rows := []string{
		m.glabel(globalFieldHeader, "header") + m.ginputs[0].View(),
		m.glabel(globalFieldPathForce, "path_force") + m.ginputs[1].View(),
		m.glabel(globalFieldEnvFile, "env_file") + m.ginputs[2].View(),
		m.glabel(globalFieldEnv, "env") + m.ginputs[3].View(),
		m.glabel(globalFieldVars, "vars") + m.ginputs[4].View(),
		m.glabel(globalFieldHide, "hide") + checkbox(m.hide),
	}
	return strings.Join(rows, "\n")
}

func (m model) label(f field, s string) string {
	label := fmt.Sprintf("%-11s ", s+":")
	if m.focus == f {
		return hotStyle.Render(label)
	}
	return dimStyle.Render(label)
}

func (m model) glabel(f globalField, s string) string {
	label := fmt.Sprintf("%-10s ", s+":")
	if m.gfocus == f {
		return hotStyle.Render(label)
	}
	return dimStyle.Render(label)
}

func (m model) projectSummary() string {
	rows := []string{
		dimStyle.Render("header: ") + valueOrDash(m.cfg.Header),
		dimStyle.Render("path_force: ") + valueOrDash(m.cfg.PathForce),
		dimStyle.Render("env_file: ") + valueOrDash(strings.Join([]string(m.cfg.EnvFile), ", ")),
		dimStyle.Render("env: ") + valueOrDash(joinStringMap(m.cfg.Env)),
		dimStyle.Render("vars: ") + valueOrDash(joinVars(m.cfg.Vars)),
		dimStyle.Render("hide: ") + fmt.Sprintf("%t", m.cfg.Hide),
	}
	return strings.Join(rows, "\n")
}

func valueOrDash(value string) string {
	if value == "" {
		return dimStyle.Render("-")
	}
	return value
}

func checkbox(v bool) string {
	if v {
		return hotStyle.Render("[x]")
	}
	return "[ ]"
}
