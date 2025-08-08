// Package input provides a common interface for UI elements in the application.
package input

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nais/cli/internal/input/button"
	"github.com/nais/cli/internal/input/confirm"
	"github.com/nais/cli/internal/input/text"
)

type Element interface {
	Focus() tea.Cmd
	Blur()
	View() string
	Valid() bool
}

func UpdateSubElement(subElement Element, msg tea.Msg) (Element, tea.Cmd) {
	switch subModel := subElement.(type) {
	case *button.Model:
		newModel, cmd := subModel.Update(msg)
		return &newModel, cmd
	case *confirm.Model:
		newModel, cmd := subModel.Update(msg)
		return &newModel, cmd
	case *text.Model:
		newModel, cmd := subModel.Update(msg)
		return &newModel, cmd
	default:
		fmt.Printf("Unhandled element type: %T\n", subElement)
		return subElement, nil
	}
}
