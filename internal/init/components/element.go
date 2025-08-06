// Package components provides a common interface for UI elements in the application.
package components

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nais/cli/internal/init/components/button"
	"github.com/nais/cli/internal/init/components/confirm"
)

type Element interface {
	Focus() tea.Cmd
	Blur()
	View() string
}

func UpdateSubElement(subElement Element, msg tea.Msg) (Element, tea.Cmd) {
	switch subModel := subElement.(type) {
	case *textinput.Model:
		newModel, cmd := subModel.Update(msg)
		return &newModel, cmd
	case *button.Model:
		newModel, cmd := subModel.Update(msg)
		return &newModel, cmd
	case *confirm.Model:
		newModel, cmd := subModel.Update(msg)
		return &newModel, cmd
	default:
		fmt.Printf("Unhandled element type: %T\n", subElement)
		return subElement, nil
	}
}
