package editor

import (
	"os"
	"path/filepath"

	"github.com/AaronKaa/mk/internal/config"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type field int

const (
	fieldName field = iota
	fieldCommand
	fieldAliases
	fieldHelp
	fieldUsage
	fieldGroup
	fieldDir
	fieldPathForce
	fieldDeps
	fieldEnvFile
	fieldEnv
	fieldVars
	fieldOpen
	fieldParallel
	fieldConfirm
	fieldHideCommand
	fieldCount
)

type globalField int

const (
	globalFieldHeader globalField = iota
	globalFieldPathForce
	globalFieldEnvFile
	globalFieldEnv
	globalFieldVars
	globalFieldHide
	globalFieldCount
)

type screen int

const (
	screenBrowse screen = iota
	screenEdit
	screenGlobal
)

var (
	panelStyle = lipgloss.NewStyle().Padding(1, 2)
	hotStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	dimStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

type model struct {
	path     string
	cfg      config.Config
	names    []string
	index    int
	screen   screen
	focus    field
	inputs   []textinput.Model
	gfocus   globalField
	ginputs  []textinput.Model
	open     bool
	parallel bool
	confirm  bool
	hide     bool
	status   string
	err      string
	quitting bool
}

func Run(args []string) error {
	path := ""
	if len(args) > 0 {
		path = args[0]
	} else if found, err := config.Find("."); err == nil {
		path = found
	} else {
		path = "mk.json"
	}

	cfg := config.Config{Commands: map[string]config.Command{}}
	if _, err := filepath.Abs(path); err == nil {
		if _, statErr := os.Stat(path); statErr == nil {
			loaded, err := config.Load(path)
			if err != nil {
				return err
			}
			cfg = loaded
		} else if !os.IsNotExist(statErr) {
			return statErr
		}
	}

	m := newModel(path, cfg)
	_, err := tea.NewProgram(m).Run()
	return err
}

func newModel(path string, cfg config.Config) model {
	if cfg.Commands == nil {
		cfg.Commands = map[string]config.Command{}
	}
	m := model{
		path:    path,
		cfg:     cfg,
		names:   config.SortedNames(cfg),
		inputs:  make([]textinput.Model, 12),
		ginputs: make([]textinput.Model, 5),
	}
	for i := range m.inputs {
		ti := textinput.New()
		ti.CharLimit = 0
		ti.Width = 72
		m.inputs[i] = ti
	}
	for i := range m.ginputs {
		ti := textinput.New()
		ti.CharLimit = 0
		ti.Width = 72
		m.ginputs[i] = ti
	}
	m.applyFocus()
	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.screen == screenEdit {
		return m.updateEdit(msg)
	}
	if m.screen == screenGlobal {
		return m.updateGlobal(msg)
	}
	return m.updateBrowse(msg)
}
