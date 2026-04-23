package editor

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) updateBrowse(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "ctrl+s", "s":
			return m.save()
		case "g":
			m.beginGlobalEdit()
			return m, nil
		case "ctrl+n", "n":
			m.addCommand()
			m.beginEdit()
			return m, nil
		case "ctrl+d", "d":
			m.deleteCurrent()
			return m, nil
		case "up", "k":
			if m.index > 0 {
				m.index--
			}
			return m, nil
		case "down", "j":
			if m.index < len(m.names)-1 {
				m.index++
			}
			return m, nil
		case "enter", "e":
			if len(m.names) == 0 {
				m.addCommand()
			}
			m.beginEdit()
			return m, nil
		}
	}
	return m, nil
}

func (m model) updateEdit(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "esc":
			m.persistCurrent()
			if m.err == "" {
				m.screen = screenBrowse
				m.blurInputs()
				m.status = "returned to command list"
			}
			return m, nil
		case "ctrl+s":
			return m.save()
		case "tab":
			m.focus = (m.focus + 1) % fieldCount
			m.applyFocus()
			return m, nil
		case "shift+tab":
			m.focus = (m.focus + fieldCount - 1) % fieldCount
			m.applyFocus()
			return m, nil
		case " ":
			if m.focus == fieldOpen {
				m.open = !m.open
				m.status = "open toggled"
				return m, nil
			}
			if m.focus == fieldParallel {
				m.parallel = !m.parallel
				m.status = "parallel toggled"
				return m, nil
			}
			if m.focus == fieldConfirm {
				m.confirm = !m.confirm
				m.status = "confirm toggled"
				return m, nil
			}
			if m.focus == fieldHideCommand {
				m.hide = !m.hide
				m.status = "hide toggled"
				return m, nil
			}
		}
	}

	if m.focus < fieldOpen {
		var cmd tea.Cmd
		m.inputs[m.focus], cmd = m.inputs[m.focus].Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m model) updateGlobal(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "esc":
			m.persistGlobal()
			if m.err == "" {
				m.screen = screenBrowse
				m.blurInputs()
				m.status = "returned to command list"
			}
			return m, nil
		case "ctrl+s":
			return m.save()
		case "tab":
			m.gfocus = (m.gfocus + 1) % globalFieldCount
			m.applyGlobalFocus()
			return m, nil
		case "shift+tab":
			m.gfocus = (m.gfocus + globalFieldCount - 1) % globalFieldCount
			m.applyGlobalFocus()
			return m, nil
		case " ":
			if m.gfocus == globalFieldHide {
				m.hide = !m.hide
				m.status = "hide toggled"
				return m, nil
			}
		}
	}

	if m.gfocus < globalFieldHide {
		var cmd tea.Cmd
		m.ginputs[m.gfocus], cmd = m.ginputs[m.gfocus].Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}
