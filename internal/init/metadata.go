package init

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nais/cli/internal/init/components"
	"github.com/nais/cli/internal/init/components/button"
	"github.com/nais/cli/internal/init/components/grid"
)

const dotChar = " • "

var (
	subtleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	dot         = lipgloss.NewStyle().Foreground(lipgloss.Color("236")).Render(dotChar)
)

type metadataModel struct {
	name   textinput.Model
	team   textinput.Model
	submit button.Model

	grid grid.Model
}

func (m *metadataModel) Init() tea.Cmd {
	name := textinput.New()
	name.Prompt = "Name: "
	name.CharLimit = 30
	name.Width = 20

	team := textinput.New()
	team.Prompt = "Team: "
	team.CharLimit = 30
	team.Width = 20

	submit := button.New("Submit")

	m.grid = grid.Model{
		Elements: [][]components.Element{
			{&name},
			{&team},
			{&submit},
		},
	}

	return m.grid.Init()
}

func (m *metadataModel) View() string {
	return m.grid.View()
}

func (m *metadataModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.grid, cmd = m.grid.Update(msg)
	return m, cmd
}
