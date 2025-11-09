package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type textInputModel struct {
	textinput textinput.Model
	label     string
	value     string
	focused   bool
}

func newTextInput(label, placeholder string) textInputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Focus()
	ti.CharLimit = 200
	ti.Width = 60
	ti.PromptStyle = lipgloss.NewStyle().Foreground(primaryColor)
	ti.TextStyle = lipgloss.NewStyle().Foreground(whiteColor)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(primaryColor)

	return textInputModel{
		textinput: ti,
		label:     label,
		focused:   true,
	}
}

func (m textInputModel) Update(msg tea.Msg) (textInputModel, tea.Cmd) {
	var cmd tea.Cmd
	m.textinput, cmd = m.textinput.Update(msg)
	m.value = m.textinput.Value()
	return m, cmd
}

func (m *textInputModel) SetValue(value string) {
	m.textinput.SetValue(value)
	m.value = value
}

func (m textInputModel) View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		labelStyle.Render(m.label),
		"",
		m.textinput.View(),
	)
}

type multiSelectModel struct {
	options  []string
	selected map[int]bool
	cursor   int
	label    string
	values   []string
}

func newMultiSelect(label string, options []string) multiSelectModel {
	return multiSelectModel{
		options:  options,
		selected: make(map[int]bool),
		cursor:   0,
		label:    label,
		values:   options,
	}
}

func (m multiSelectModel) Update(msg tea.Msg) (multiSelectModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case " ", "x":
			m.selected[m.cursor] = !m.selected[m.cursor]
		}
	}
	return m, nil
}

func (m multiSelectModel) View() string {
	var items []string
	items = append(items, labelStyle.Render(m.label))
	items = append(items, "")

	for i, option := range m.options {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checkbox := "[ ]"
		if m.selected[i] {
			checkbox = checkboxStyle.Render("[✓]")
		}

		style := unselectedStyle
		if m.cursor == i {
			style = selectedStyle
		}

		items = append(items, style.Render(cursor+" "+checkbox+" "+option))
	}

	return lipgloss.JoinVertical(lipgloss.Left, items...)
}

func (m multiSelectModel) GetSelected() []string {
	var selected []string
	for i, option := range m.options {
		if m.selected[i] {
			selected = append(selected, option)
		}
	}
	return selected
}

type singleSelectModel struct {
	options  []string
	selected int
	cursor   int
	label    string
	values   []string
}

func newSingleSelect(label string, options []string, defaultIndex int) singleSelectModel {
	return singleSelectModel{
		options:  options,
		selected: defaultIndex,
		cursor:   0,
		label:    label,
		values:   options,
	}
}

func (m singleSelectModel) Update(msg tea.Msg) (singleSelectModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case " ", "enter":
			m.selected = m.cursor
		}
	}
	return m, nil
}

func (m singleSelectModel) View() string {
	var items []string
	items = append(items, labelStyle.Render(m.label))
	items = append(items, "")

	for i, option := range m.options {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		radio := "( )"
		if m.selected == i {
			radio = checkboxStyle.Render("(●)")
		}

		style := unselectedStyle
		if m.cursor == i {
			style = selectedStyle
		}

		items = append(items, style.Render(cursor+" "+radio+" "+option))
	}

	return lipgloss.JoinVertical(lipgloss.Left, items...)
}

func (m singleSelectModel) GetSelected() string {
	if m.selected >= 0 && m.selected < len(m.values) {
		return m.values[m.selected]
	}
	return ""
}
