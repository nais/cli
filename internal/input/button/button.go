// Package button provides a simple button component for the Bubble Tea framework.
package button

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
)

type Model struct {
	focused bool
	label   string
}

func New(label string) Model {
	return Model{
		focused: false,
		label:   label,
	}
}

func (m *Model) Focus() tea.Cmd {
	m.focused = true
	return nil
}

func (m *Model) Blur() {
	m.focused = false
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m Model) View() string {
	out := "[ " + m.label + " ]"
	if m.focused {
		return focusedStyle.Render(out)
	}
	return blurredStyle.Render(out)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (Model) Valid() bool {
	return true
}
