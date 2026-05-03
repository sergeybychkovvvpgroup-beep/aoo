package ui

import (
	"fmt"
	"sort"
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
	tags      []string
	tagIndex  int
	options   Options
	cursor    int
	width     int
	height    int
	selected  *notes.Entry
	edit      bool
	cancelled bool
	theme     Theme
}

type doneMsg struct{}

func NewPicker(entries []notes.Entry, initialQuery string, theme Theme, options Options) PickerModel {
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
		input:   input,
		entries: entries,
		tags:    availableTags(entries),
		theme:   theme,
		options: options,
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
		if inputWidth := m.inputWidth(); inputWidth > 8 {
			m.input.Width = inputWidth
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
		case "alt+e":
			if len(m.matches) == 0 {
				return m, nil
			}
			entry := m.matches[m.cursor].Entry
			m.selected = &entry
			m.edit = true
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
			if len(m.tags) > 1 {
				m.tagIndex--
				if m.tagIndex < 0 {
					m.tagIndex = len(m.tags) - 1
				}
				m.refresh()
			}
			return m, nil
		case "right":
			if len(m.tags) > 1 {
				m.tagIndex = (m.tagIndex + 1) % len(m.tags)
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
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.TitleFG)).Bold(true)
	inputBox := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.InputFG)).
		Background(lipgloss.Color(m.theme.InputBG)).
		Padding(0, 0)
	dividerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.DividerFG))

	var resultLines []string
	listWidth, previewWidth := m.layoutWidths()
	mainPaneStyle := lipgloss.NewStyle().Width(listWidth).MaxWidth(listWidth)
	contentWidth := listWidth

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
			if m.options.ShowPreview {
				detailText := truncateRunes(match.Detail, maxInt(10, contentWidth-4))
				resultLines = append(resultLines, "    "+detailStyle.Render(detailText))
				if i != len(visible)-1 {
					resultLines = append(resultLines, "    "+dividerStyle.Render("·"))
				}
			}
		}
	}

	mainLines := make([]string, 0, len(resultLines)+9)
	mainLines = append(mainLines, "")
	mainLines = append(mainLines, resultLines...)
	footerLines := []string{
		"",
		inputBox.Render(m.input.View()),
		titleStyle.Render(m.statusLine()),
		helpStyle.Render("enter select  alt+e edit  esc quit  ←→ filter  ↑↓ move"),
	}

	effectiveHeight := m.effectiveHeight()
	if effectiveHeight > 0 {
		used := 1 + len(resultLines) + len(footerLines)
		if filler := effectiveHeight - used; filler > 0 {
			for i := 0; i < filler; i++ {
				mainLines = append(mainLines, "")
			}
		}
	} else {
		mainLines = append(mainLines, "")
	}

	mainLines = append(mainLines, footerLines...)
	mainView := mainPaneStyle.Render(strings.Join(mainLines, "\n"))

	if !m.shouldShowPreviewPane(previewWidth) {
		return mainView
	}

	previewText := m.previewContent(previewWidth - 4)
	previewHeight := maxInt(effectiveHeight, lipgloss.Height(mainView))
	previewView := lipgloss.NewStyle().
		Width(previewWidth-2).
		Height(maxInt(3, previewHeight)).
		Padding(0, 1).
		Render(previewText)

	separator := dividerStyle.Render(strings.Repeat("│\n", maxInt(1, previewHeight)))
	separator = strings.TrimRight(separator, "\n")
	return lipgloss.JoinHorizontal(lipgloss.Top, mainView, " ", separator, " ", previewView)
}

func (m PickerModel) Selected() *notes.Entry {
	return m.selected
}

func (m PickerModel) Cancelled() bool {
	return m.cancelled
}

func (m PickerModel) EditRequested() bool {
	return m.edit
}

func (m PickerModel) Query() string {
	return m.input.Value()
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
	tag := m.activeTag()
	if tag == notes.TypeAll {
		return m.entries
	}

	filtered := make([]notes.Entry, 0, len(m.entries))
	for _, entry := range m.entries {
		for _, entryTag := range entry.Tags {
			if entryTag == tag {
				filtered = append(filtered, entry)
				break
			}
		}
	}
	return filtered
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
	available := m.effectiveHeight() - 5
	if available < 6 {
		available = 6
	}

	if m.options.ShowPreview {
		// One item uses 2 lines, plus 1 separator between items => roughly 3n-1 lines.
		maxItems := (available + 1) / 3
		if maxItems < 2 {
			maxItems = 2
		}
		return maxItems
	}

	maxItems := available
	if maxItems < 3 {
		maxItems = 3
	}
	return maxItems
}

func RunPicker(entries []notes.Entry, initialQuery string, themeName string, options Options) (*notes.Entry, string, bool, bool, error) {
	theme, err := ResolveTheme(themeName)
	if err != nil {
		return nil, initialQuery, false, false, err
	}

	model := NewPicker(entries, initialQuery, theme, options)
	programOptions := []tea.ProgramOption{}
	if options.FullScreen {
		programOptions = append(programOptions, tea.WithAltScreen())
	}
	program := tea.NewProgram(model, programOptions...)
	result, err := program.Run()
	if err != nil {
		return nil, initialQuery, false, false, fmt.Errorf("picker: %w", err)
	}

	finalModel, ok := result.(PickerModel)
	if !ok {
		return nil, initialQuery, false, false, fmt.Errorf("picker returned unexpected model type")
	}

	return finalModel.Selected(), finalModel.Query(), finalModel.Cancelled(), finalModel.EditRequested(), nil
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

func (m PickerModel) effectiveHeight() int {
	height := m.height
	if height <= 0 {
		height = m.options.Height
	}
	if m.options.Height > 0 && (height <= 0 || m.options.Height < height) {
		height = m.options.Height
	}
	if height < 6 {
		return 6
	}
	return height
}

func (m PickerModel) layoutWidths() (int, int) {
	total := m.contentWidth()
	if !m.options.PreviewPane || total < 80 {
		return total, 0
	}

	previewWidth := total / 3
	if previewWidth < 32 {
		previewWidth = 32
	}
	if previewWidth > 52 {
		previewWidth = 52
	}

	listWidth := total - previewWidth - 3
	if listWidth < 24 {
		return total, 0
	}
	return listWidth, previewWidth
}

func (m PickerModel) inputWidth() int {
	listWidth, _ := m.layoutWidths()
	if listWidth <= 8 {
		return 48
	}
	return listWidth - 4
}

func (m PickerModel) activeTag() string {
	if len(m.tags) == 0 {
		return notes.TypeAll
	}
	if m.tagIndex < 0 || m.tagIndex >= len(m.tags) {
		return notes.TypeAll
	}
	return m.tags[m.tagIndex]
}

func availableTags(entries []notes.Entry) []string {
	seen := map[string]struct{}{
		notes.TypeAll: {},
	}
	tags := []string{notes.TypeAll}
	var collected []string

	for _, entry := range entries {
		for _, tag := range entry.Tags {
			tag = strings.TrimSpace(strings.ToLower(tag))
			if tag == "" {
				continue
			}
			if _, ok := seen[tag]; ok {
				continue
			}
			seen[tag] = struct{}{}
			collected = append(collected, tag)
		}
	}

	sort.Strings(collected)
	return append(tags, collected...)
}

func (m PickerModel) shouldShowPreviewPane(previewWidth int) bool {
	return m.options.PreviewPane && previewWidth > 0
}

func (m PickerModel) previewContent(width int) string {
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.TitleFG)).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.HelpFG))
	bodyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.DetailFG))

	if len(m.matches) == 0 {
		return bodyStyle.Render("No selection")
	}

	entry := m.matches[m.cursor].Entry
	meta := make([]string, 0, 3)
	if entry.SourceFile != "" {
		meta = append(meta, entry.SourceFile)
	}
	if action := entry.Action(); strings.TrimSpace(action) != "" {
		meta = append(meta, strings.ToLower(action))
	}
	if tag := m.activeTag(); tag != notes.TypeAll {
		meta = append(meta, "tag:"+tag)
	}
	lines := []string{titleStyle.Render(truncateRunes(entry.Desc, width))}
	if len(meta) > 0 {
		lines = append(lines, labelStyle.Render(truncateRunes(strings.Join(meta, "  ·  "), width)))
	}

	if entry.HasShow() {
		lines = append(lines, "")
		if action := entry.FirstShowAction(); action != nil {
			lines = append(lines, wrapText(strings.TrimSpace(action.Text), width)...)
		}
	}

	if entry.HasCmd() {
		lines = append(lines, "", labelStyle.Render("commands"))
		for i, action := range entry.CmdActions() {
			label := strings.TrimSpace(action.Desc)
			if label == "" {
				label = fmt.Sprintf("command %d", i+1)
			}
			lines = append(lines, bodyStyle.Render(truncateRunes(label, maxInt(10, width))))
			lines = append(lines, wrapText(action.Cmd, maxInt(10, width))...)
		}
	}

	if entry.HasTemplate() {
		lines = append(lines, "", labelStyle.Render("template"))
		for _, action := range entry.ActionsList() {
			if action.IsTemplate() {
				lines = append(lines, wrapText(strings.TrimSpace(action.Template), width)...)
				break
			}
		}
	}

	return bodyStyle.Render(strings.Join(lines, "\n"))
}

func (m PickerModel) statusLine() string {
	total := len(m.filteredEntries())
	matches := len(m.matches)
	active := m.activeTag()
	if active == notes.TypeAll {
		active = "all"
	}
	return fmt.Sprintf("%d/%d  tag:%s", matches, total, active)
}

func wrapText(value string, width int) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return []string{""}
	}

	var lines []string
	for _, rawLine := range strings.Split(value, "\n") {
		rawLine = strings.TrimRight(rawLine, " ")
		if rawLine == "" {
			lines = append(lines, "")
			continue
		}
		runes := []rune(rawLine)
		for len(runes) > width && width > 1 {
			lines = append(lines, string(runes[:width]))
			runes = runes[width:]
		}
		lines = append(lines, string(runes))
	}
	return lines
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
