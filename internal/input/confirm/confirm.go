// Package confirm provides a simple confirmation prompt
package confirm

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
)

type Model struct {
	prompt  string
	focused bool
	answer  bool
}

func New(prompt string, defaultAnswer bool) Model {
	return Model{
		prompt:  prompt,
		focused: false,
		answer:  defaultAnswer,
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

func (m Model) viewButtons() string {
	no := "[ no ]"
	yes := "[ yes ]"
	if !m.focused {
		return blurredStyle.Render(no + " " + yes)
	}

	if m.answer {
		return blurredStyle.Render(no) + " " + focusedStyle.Render(yes)
	}
	return focusedStyle.Render(no) + " " + blurredStyle.Render(yes)
}

func (m Model) View() string {
	return m.prompt + "\n" + m.viewButtons()
}

func (m *Model) Answer() bool {
	return m.answer
}

func (m Model) FocusNextButton() (Model, tea.Cmd) {
	m.answer = !m.answer
	return m, nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "right", "tab", "shift+tab":
			return m.FocusNextButton()
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			return m, nil
		}
	}
	return m, nil
}
