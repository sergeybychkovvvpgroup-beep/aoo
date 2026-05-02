package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sergeyb/aoo/internal/notes"
)

type PickerModel struct {
	input     textinput.Model
	entries   []notes.Entry
	matches   []notes.Match
	cursor    int
	width     int
	height    int
	selected  *notes.Entry
	cancelled bool
}

type doneMsg struct{}

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	detailStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

func NewPicker(entries []notes.Entry, initialQuery string) PickerModel {
	input := textinput.New()
	input.Placeholder = "Search notes"
	input.Prompt = "notes> "
	input.SetValue(initialQuery)
	input.Focus()
	input.CharLimit = 256
	input.Width = 64

	m := PickerModel{
		input:   input,
		entries: entries,
	}
	m.refresh()
	return m
}

func (m PickerModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m PickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			if len(m.matches) == 0 {
				return m, nil
			}
			entry := m.matches[m.cursor].Entry
			m.selected = &entry
			return m, tea.Quit
		case "up", "ctrl+k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "ctrl+j":
			if m.cursor < len(m.matches)-1 {
				m.cursor++
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.refresh()
	return m, cmd
}

func (m PickerModel) View() string {
	var lines []string
	lines = append(lines, titleStyle.Render("aoo"))
	lines = append(lines, m.input.View())
	lines = append(lines, "")

	if len(m.matches) == 0 {
		lines = append(lines, detailStyle.Render("No matches"))
	} else {
		for i, match := range m.visibleMatches() {
			prefix := "  "
			label := match.Label
			if i+m.offset() == m.cursor {
				prefix = "> "
				label = selectedStyle.Render(label)
			}
			lines = append(lines, prefix+label)
			lines = append(lines, "    "+detailStyle.Render(match.Detail))
		}
	}

	lines = append(lines, "")
	lines = append(lines, helpStyle.Render("Enter: open/run  Esc: quit  Up/Down: move"))
	return strings.Join(lines, "\n")
}

func (m PickerModel) Selected() *notes.Entry {
	return m.selected
}

func (m PickerModel) Cancelled() bool {
	return m.cancelled
}

func (m *PickerModel) refresh() {
	m.matches = notes.Filter(m.entries, m.input.Value())
	if m.cursor >= len(m.matches) {
		m.cursor = len(m.matches) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m PickerModel) visibleMatches() []notes.Match {
	if len(m.matches) == 0 {
		return nil
	}

	available := m.height - 6
	if available < 3 {
		available = 6
	}

	maxItems := available / 2
	if maxItems < 5 {
		maxItems = 5
	}

	start := m.offset()
	end := start + maxItems
	if end > len(m.matches) {
		end = len(m.matches)
	}

	return m.matches[start:end]
}

func (m PickerModel) offset() int {
	if len(m.matches) == 0 {
		return 0
	}
	window := 5
	if m.height > 0 {
		window = max(5, (m.height-6)/2)
	}

	if m.cursor < window/2 {
		return 0
	}

	start := m.cursor - window/2
	limit := len(m.matches) - window
	if limit < 0 {
		return 0
	}
	if start > limit {
		return limit
	}
	return start
}

func RunPicker(entries []notes.Entry, initialQuery string) (*notes.Entry, bool, error) {
	model := NewPicker(entries, initialQuery)
	program := tea.NewProgram(model, tea.WithAltScreen())
	result, err := program.Run()
	if err != nil {
		return nil, false, fmt.Errorf("picker: %w", err)
	}

	finalModel, ok := result.(PickerModel)
	if !ok {
		return nil, false, fmt.Errorf("picker returned unexpected model type")
	}

	return finalModel.Selected(), finalModel.Cancelled(), nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
