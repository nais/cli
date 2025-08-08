package text

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))

type Model struct {
	textinput.Model
}

type Option func(*Model)

func WithValidation(fn textinput.ValidateFunc) Option {
	return func(m *Model) {
		m.Validate = fn
	}
}

func New(prompt, placeholder string, opts ...Option) Model {
	model := textinput.New()
	model.Prompt = strings.TrimSpace(prompt) + " "
	model.Placeholder = strings.TrimSpace(placeholder)
	model.CharLimit = 30
	model.Width = 20

	m := Model{Model: model}

	for _, opt := range opts {
		opt(&m)
	}

	return m
}

func (m Model) View() string {
	ret := ""
	if m.Err != nil {
		ret = errorStyle.Render("Validation error: "+m.Err.Error()) + "\n"
	}

	return ret + m.Model.View()
}

func (m Model) Valid() bool {
	if m.Validate == nil {
		return true
	}
	return m.Validate(m.Value()) == nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	newModel, cmd := m.Model.Update(msg)
	m.Model = newModel
	return m, cmd
}
