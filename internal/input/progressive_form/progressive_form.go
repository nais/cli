// Package progressiveform provides a way to progressively fill out a form in a terminal UI.
package progressiveform

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nais/cli/internal/input"
)

const dotChar = " • "

var (
	subtleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	dot         = lipgloss.NewStyle().Foreground(lipgloss.Color("236")).Render(dotChar)
)

type Model struct {
	currentElement int
	Elements       []input.Element
}

func (m *Model) Init() tea.Cmd {
	m.activeElement().Focus()
	return nil
}

func (m *Model) View() string {
	out := &strings.Builder{}
	for i, element := range m.Elements {
		out.WriteString(strings.TrimSpace((element.View())))
		if m.currentElement == i {
			break
		} else {
			out.WriteString(" ✓")
		}
		out.WriteRune('\n')
	}
	if m.currentElement == len(m.Elements) {
		out.WriteString("✓ All elements completed!\n")
	}

	out.WriteRune('\n')
	out.WriteRune('\n')
	// out.WriteString(subtleStyle.Render("[shift] tab: move focus") + dot)
	out.WriteString(subtleStyle.Render("enter: submit") + dot)
	// out.WriteString(subtleStyle.Render("space: select") + dot)
	out.WriteString(subtleStyle.Render("esc: quit"))

	return out.String()
}

func (m *Model) nextElement() tea.Cmd {
	m.activeElement().Blur()

	if m.currentElement < len(m.Elements) {
		m.currentElement++
	}

	if m.currentElement < len(m.Elements) {
		return m.activeElement().Focus()
	}

	return nil
}

func (m *Model) activeElement() input.Element {
	return m.Elements[m.currentElement]
}

func (m *Model) replaceActiveElement(e input.Element) {
	m.Elements[m.currentElement] = e
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			if m.currentElement > len(m.Elements)-1 {
				return m, tea.Quit
			}
			if m.activeElement().Valid() {
				return m, m.nextElement()
			}
		}
	}

	if m.currentElement >= len(m.Elements) {
		// done
		return m, nil
	}

	updatedElement, cmd := input.UpdateSubElement(m.activeElement(), msg)
	m.replaceActiveElement(updatedElement)
	return m, cmd
}

func (Model) Valid() bool {
	return true
}
