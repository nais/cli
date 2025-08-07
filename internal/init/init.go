// Package init is a command to get started with a new nais application or naisjob
package init

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

import (
	// nais_io_v1 "github.com/nais/liberator/pkg/apis/nais.io/v1"
	// nais_io_v1alpha1 "github.com/nais/liberator/pkg/apis/nais.io/v1alpha1"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nais/cli/internal/init/command/flag"
	"github.com/nais/naistrix"
)

type model struct {
	models      map[string]tea.Model
	activeModel string
}

func initialModel(flags *flag.Init) model {
	return model{
		models: map[string]tea.Model{
			"metadata": &metadataFlowModel{
				flags: flags,
			},
		},
		activeModel: "metadata",
	}
}

func Run(flags *flag.Init, out naistrix.Output) error {
	m := initialModel(flags)
	if _, err := tea.NewProgram(&m).Run(); err != nil {
		out.Printf("Could not start nais init prompter: %s\n", err)
		return err
	}

	return nil
}

func (m *model) Init() tea.Cmd {
	cmds := []tea.Cmd{}
	for _, model := range m.models {
		cmd := model.Init()
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}

func (m *model) View() string {
	return m.models[m.activeModel].View()
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.models[m.activeModel].Update(msg)
}
