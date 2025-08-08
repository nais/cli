// Package grid provides a grid layout for interactive elements in a terminal UI.
package grid

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nais/cli/internal/init/components/button"
)

const dotChar = " • "

var (
	subtleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	dot         = lipgloss.NewStyle().Foreground(lipgloss.Color("236")).Render(dotChar)
)

type Model struct {
	row, col int
	Elements [][]Element
}

type Element interface {
	Focus() tea.Cmd
	Blur()
	View() string
}

func (m *Model) Init() tea.Cmd {
	m.activeElement().Focus()
	return nil
}

func (m *Model) View() string {
	out := &strings.Builder{}
	for _, row := range m.Elements {
		for _, element := range row {
			out.WriteString(element.View())
		}
		out.WriteRune('\n')
	}

	out.WriteString(subtleStyle.Render("[shift] tab: move focus") + dot)
	out.WriteString(subtleStyle.Render("enter: submit") + dot)
	out.WriteString(subtleStyle.Render("space: select") + dot)
	out.WriteString(subtleStyle.Render("esc: quit"))

	return out.String()
}

func (m *Model) nextElement() tea.Cmd {
	m.activeElement().Blur()

	if m.col+1 < len(m.Elements[m.row]) {
		m.col++
	} else if m.row+1 < len(m.Elements) {
		m.row++
		m.col = 0
	}

	return m.activeElement().Focus()
}

func (m *Model) previousElement() tea.Cmd {
	m.activeElement().Blur()

	if m.col-1 >= 0 {
		m.col--
	} else if m.row-1 >= 0 {
		m.row--
		m.col = 0
	}

	return m.activeElement().Focus()
}

func (m *Model) activeElement() Element {
	return m.Elements[m.row][m.col]
}

func (m *Model) replaceActiveElement(e Element) {
	m.Elements[m.row][m.col] = e
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m, nil
		case "tab":
			return m, m.nextElement()
		case "shift+tab":
			return m, m.previousElement()
		}
	}

	switch element := m.activeElement().(type) {
	case *textinput.Model:
		updatedElement, cmd := element.Update(msg)
		m.replaceActiveElement(&updatedElement)
		return m, cmd
	case *button.Model:
		updatedElement, cmd := element.Update(msg)
		m.replaceActiveElement(&updatedElement)
		return m, cmd
	default:
		fmt.Printf("Unhandled element type: %T\n", element)
		return m, nil
	}
}
