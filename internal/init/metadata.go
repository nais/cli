package init

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type metadataModel struct {
	focusIndex int
	inputs     []textinput.Model
}

func (m *metadataModel) Init() tea.Cmd {
	name := textinput.New()
	name.Placeholder = "App Name"
	name.CharLimit = 30
	name.Width = 20

	team := textinput.New()
	team.Placeholder = "Team Name"
	team.CharLimit = 30
	team.Width = 20

	m.focusIndex = 0
	m.inputs = append(m.inputs, name, team)
	m.inputs[m.focusIndex].Focus()

	return nil
}

func (m *metadataModel) View() string {
	out := &strings.Builder{}
	for _, input := range m.inputs {
		out.WriteString(input.View())
		out.WriteRune('\n')
	}

	return out.String()
}

func (m *metadataModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.inputs[m.focusIndex].Blur()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter", "tab", "down":
			m.focusIndex++
		case "shift+tab", "up":
			m.focusIndex--
		}
	}

	m.focusIndex = m.focusIndex % len(m.inputs)
	m.inputs[m.focusIndex].Focus()

	var cmd tea.Cmd
	m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)

	return m, cmd
}
