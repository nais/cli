package init

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nais/cli/internal/init/components/button"
	progressiveform "github.com/nais/cli/internal/init/components/progressive_form"
)

type metadataFlowModel struct {
	name   textinput.Model
	team   textinput.Model
	submit button.Model

	flow progressiveform.Model
}

func (m *metadataFlowModel) Init() tea.Cmd {
	name := textinput.New()
	name.Prompt = "What should the app's name be?\n"
	name.Placeholder = "name"
	name.CharLimit = 30
	name.Width = 20

	team := textinput.New()
	team.Prompt = "In which team should this app live?\n"
	team.Placeholder = "team"
	team.CharLimit = 30
	team.Width = 20

	m.flow = progressiveform.Model{
		Elements: []progressiveform.Element{
			&name,
			&team,
		},
	}

	return m.flow.Init()
}

func (m *metadataFlowModel) View() string {
	return m.flow.View()
}

func (m *metadataFlowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.flow, cmd = m.flow.Update(msg)
	return m, cmd
}
