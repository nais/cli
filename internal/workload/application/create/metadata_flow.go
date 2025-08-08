package create

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nais/cli/internal/input"
	"github.com/nais/cli/internal/input/button"
	"github.com/nais/cli/internal/input/confirm"
	progressiveform "github.com/nais/cli/internal/input/progressive_form"
	"github.com/nais/cli/internal/input/text"
	"github.com/nais/cli/internal/workload/application/command/flag"
)

type metadataFlowModel struct {
	name   textinput.Model
	team   textinput.Model
	submit button.Model
	flags  *flag.Create

	flow progressiveform.Model
}

/*
Ressurser og skalering
- Hvordan vil du skalere appen?
- Hvor mye minne/cpu vil den trenge?
- Ender i ferdig konfigurert resources og replicas

Kommunikasjon
- Er det noen som skal kommunisere med appen din? mennesker? maskiner? begge? internt / ekstern?
- Skal appen din kommunisere med andre? tokens? internt? eksternt? endepunkter?
Disse spørsmålene kan struktureres i ett tre, og bør ende opp i ferdig konfigurasjon for bl.a. auth, ingresses og accessPolicies.

Persistens
- postgres, kafka, bucket, bq, valkey, opensearch?

Defaults
Bruker den nais standarder? port: 8080, metrikker på /metrics, isalive på /isalive, isready etc?
+Hemmeligheter? +Miljøvariabler?
*/
func (m *metadataFlowModel) Init() tea.Cmd {
	name := text.New("Provide a name for the application:", "name", text.WithValidation(func(s string) error {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("application name can not be empty")
		}

		return nil
	}))

	if m.flags.Name != "" {
		name.SetValue(m.flags.Name)
	}

	team := text.New("Specify the team that owns the application:", "team")

	if m.flags.Team != "" {
		team.SetValue(m.flags.Team)
	}

	scaling := confirm.New("Should the app be automatically scaled?", true)

	// replicas := textinput.New()
	// replicas.Prompt = "How many replicas should the app have before scaling?"
	// replicas.Placeholder = ""
	// replicas.CharLimit = 2
	// replicas.Width = 2
	// replicas.Validate = func(s string) error {
	// 	return nil
	// }

	m.flow = progressiveform.Model{
		Elements: []input.Element{
			&name,
			&team,
			&scaling,
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
