package ui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"aoo/internal/notes"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PickerModel struct {
	input     textinput.Model
	entries   []notes.Entry
	matches   []notes.Match
	noteTypes []string
	typeIndex int
	cursor    int
	width     int
	height    int
	selected  *notes.Entry
	cancelled bool
	theme     Theme
}

type doneMsg struct{}

func NewPicker(entries []notes.Entry, initialQuery string, theme Theme) PickerModel {
	input := textinput.New()
	input.Placeholder = "search"
	input.Prompt = "notes> "
	input.SetValue(initialQuery)
	input.Focus()
	input.CharLimit = 256
	input.Width = 48
	input.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.InputFG))
	input.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.InputPrompt))
	input.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(theme.HelpFG))

	m := PickerModel{
		input:     input,
		entries:   entries,
		noteTypes: availableNoteTypes(entries),
		theme:     theme,
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
		if m.width > 8 {
			m.input.Width = m.width - 8
		}
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
		case "left":
			if len(m.noteTypes) > 1 {
				m.typeIndex--
				if m.typeIndex < 0 {
					m.typeIndex = len(m.noteTypes) - 1
				}
				m.refresh()
			}
			return m, nil
		case "right":
			if len(m.noteTypes) > 1 {
				m.typeIndex = (m.typeIndex + 1) % len(m.noteTypes)
				m.refresh()
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
	rowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.RowFG))
	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.SelectedFG)).
		Background(lipgloss.Color(m.theme.SelectedBG)).
		Bold(true)
	detailStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.DetailFG))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.HelpFG))
	inputBox := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.InputFG)).
		Background(lipgloss.Color(m.theme.InputBG)).
		Padding(0, 0)
	dividerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.DividerFG))

	headerLines := []string{m.renderTypeTabs(), ""}

	var resultLines []string
	contentWidth := m.contentWidth()

	if len(m.matches) == 0 {
		resultLines = append(resultLines, detailStyle.Render("No matches"))
	} else {
		visible := m.visibleMatches()
		for i, match := range visible {
			prefix := "  "
			labelText := truncateRunes(match.Label, contentWidth-3)
			label := rowStyle.Render(labelText)
			if i+m.offset() == m.cursor {
				prefix = m.theme.SelectedMark + " "
				label = selectedStyle.Render(padRight(labelText, contentWidth-3))
			}
			resultLines = append(resultLines, prefix+label)
			detailText := truncateRunes(match.Detail, maxInt(10, contentWidth-4))
			resultLines = append(resultLines, "    "+detailStyle.Render(detailText))
			if i != len(visible)-1 {
				resultLines = append(resultLines, "    "+dividerStyle.Render("·"))
			}
		}
	}

	footerLines := []string{
		"",
		inputBox.Render(m.input.View()),
		helpStyle.Render("enter open/run   esc quit   ←→ type   ↑↓ move"),
	}

	lines := make([]string, 0, len(headerLines)+len(resultLines)+len(footerLines)+8)
	lines = append(lines, headerLines...)
	lines = append(lines, resultLines...)

	if m.height > 0 {
		used := len(headerLines) + len(resultLines) + len(footerLines)
		if filler := m.height - used; filler > 0 {
			for i := 0; i < filler; i++ {
				lines = append(lines, "")
			}
		}
	} else {
		lines = append(lines, "")
	}

	lines = append(lines, footerLines...)
	return strings.Join(lines, "\n")
}

func (m PickerModel) Selected() *notes.Entry {
	return m.selected
}

func (m PickerModel) Cancelled() bool {
	return m.cancelled
}

func (m *PickerModel) refresh() {
	m.matches = notes.Filter(m.filteredEntries(), m.input.Value())
	if m.cursor >= len(m.matches) {
		m.cursor = len(m.matches) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m PickerModel) filteredEntries() []notes.Entry {
	activeType := m.activeType()
	if activeType == notes.TypeAll {
		return m.entries
	}

	filtered := make([]notes.Entry, 0, len(m.entries))
	for _, entry := range m.entries {
		if entry.Type() == activeType {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func (m PickerModel) activeType() string {
	if len(m.noteTypes) == 0 {
		return notes.TypeAll
	}
	if m.typeIndex < 0 || m.typeIndex >= len(m.noteTypes) {
		return notes.TypeAll
	}
	return m.noteTypes[m.typeIndex]
}

func (m PickerModel) renderTypeTabs() string {
	baseStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.HelpFG))
	activeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.SelectedFG)).
		Background(lipgloss.Color(m.theme.SelectedBG)).
		Bold(true).
		Padding(0, 1)

	parts := make([]string, 0, len(m.noteTypes))
	activeType := m.activeType()
	for _, noteType := range m.noteTypes {
		label := noteType
		if noteType == activeType {
			parts = append(parts, activeStyle.Render(label))
			continue
		}
		parts = append(parts, baseStyle.Render(" "+label+" "))
	}
	return strings.Join(parts, " ")
}

func availableNoteTypes(entries []notes.Entry) []string {
	types := []string{notes.TypeAll}
	seen := map[string]struct{}{
		notes.TypeAll: {},
	}

	appendType := func(noteType string) {
		if _, exists := seen[noteType]; exists {
			return
		}
		seen[noteType] = struct{}{}
		types = append(types, noteType)
	}

	for _, noteType := range []string{notes.TypeRun, notes.TypeShow} {
		for _, entry := range entries {
			if entry.Type() == noteType {
				appendType(noteType)
				break
			}
		}
	}

	return types
}

func (m PickerModel) visibleMatches() []notes.Match {
	if len(m.matches) == 0 {
		return nil
	}

	maxItems := m.maxVisibleItems()

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
	window := m.maxVisibleItems()

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

func (m PickerModel) maxVisibleItems() int {
	available := m.height - 7
	if available < 6 {
		available = 6
	}

	// One item uses:
	// 1 line for title
	// 1 line for detail
	// 1 separator line between items
	// => n items take roughly 3n-1 lines.
	maxItems := (available + 1) / 3
	if maxItems < 2 {
		maxItems = 2
	}
	return maxItems
}

func RunPicker(entries []notes.Entry, initialQuery string, themeName string) (*notes.Entry, bool, error) {
	theme, err := ResolveTheme(themeName)
	if err != nil {
		return nil, false, err
	}

	model := NewPicker(entries, initialQuery, theme)
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

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m PickerModel) contentWidth() int {
	if m.width <= 0 {
		return 80
	}
	if m.width < 20 {
		return 20
	}
	return m.width - 4
}

func truncateRunes(value string, limit int) string {
	if limit <= 0 {
		return ""
	}
	if utf8.RuneCountInString(value) <= limit {
		return value
	}
	if limit <= 1 {
		return "…"
	}

	runes := []rune(value)
	return string(runes[:limit-1]) + "…"
}

func padRight(value string, width int) string {
	length := utf8.RuneCountInString(value)
	if length >= width {
		return value
	}
	return value + strings.Repeat(" ", width-length)
}
