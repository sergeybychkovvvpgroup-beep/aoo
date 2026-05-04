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
	input        textinput.Model
	entries      []notes.Entry
	matches      []notes.Match
	preview      notes.PreviewMatch
	tags         []string
	tagIndex     int
	options      Options
	cursor       int
	width        int
	height       int
	selected     *notes.Entry
	selectedLine int
	edit         bool
	createKind   string
	previewHit   int
	cancelled    bool
	previewCache map[string]notes.PreviewMatch
	theme        Theme
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
		input:        input,
		entries:      entries,
		tags:         availableTags(entries),
		theme:        theme,
		options:      options,
		previewCache: make(map[string]notes.PreviewMatch),
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
		case "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		case "esc":
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			if len(m.matches) == 0 {
				return m, nil
			}
			entry := m.matches[m.cursor].Entry
			m.selected = &entry
			m.selectedLine = entry.PreviewHitLine(m.preview, m.activePreviewHit())
			return m, tea.Quit
		case "ctrl+e", "alt+e":
			if len(m.matches) == 0 {
				return m, nil
			}
			entry := m.matches[m.cursor].Entry
			m.selected = &entry
			m.selectedLine = entry.PreviewHitLine(m.preview, m.activePreviewHit())
			m.edit = true
			return m, tea.Quit
		case "ctrl+n", "alt+n":
			m.createKind = "note"
			return m, tea.Quit
		case "up", "ctrl+k":
			if m.cursor > 0 {
				m.cursor--
				m.previewHit = 0
				m.refreshPreview()
			}
			return m, nil
		case "down", "ctrl+j":
			if m.cursor < len(m.matches)-1 {
				m.cursor++
				m.previewHit = 0
				m.refreshPreview()
			}
			return m, nil
		case "right":
			m.advancePreviewHit(1)
			return m, nil
		case "left":
			m.advancePreviewHit(-1)
			return m, nil
		case "f3":
			if m.cursor < len(m.matches)-1 {
				m.cursor++
				m.previewHit = 0
				m.refreshPreview()
			}
			return m, nil
		case "shift+f3":
			if m.cursor > 0 {
				m.cursor--
				m.previewHit = 0
				m.refreshPreview()
			}
			return m, nil
		case "tab":
			if len(m.tags) > 1 {
				m.tagIndex = (m.tagIndex + 1) % len(m.tags)
				m.refresh()
			}
			return m, nil
		case "shift+tab":
			if len(m.tags) > 1 {
				m.tagIndex--
				if m.tagIndex < 0 {
					m.tagIndex = len(m.tags) - 1
				}
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
	containerStyle := lipgloss.NewStyle().Width(m.contentWidth()).MaxWidth(m.contentWidth())
	contentWidth := m.contentWidth()
	effectiveHeight := m.effectiveHeight()

	lines := []string{inputBox.Render(m.input.View())}
	if strings.TrimSpace(m.input.Value()) != "" {
		lines = append(lines, "")
		lines = append(lines, m.resultLines(contentWidth, rowStyle, selectedStyle, detailStyle)...)
		lines = append(lines, "")
		lines = append(lines, titleStyle.Render(m.statusLine()))
	}
	lines = append(lines, helpStyle.Render("enter open  ctrl+n new  ctrl+e edit  esc quit"))

	if effectiveHeight > 0 && len(lines) < effectiveHeight {
		fillerAt := len(lines) - 1
		padding := make([]string, 0, effectiveHeight-len(lines))
		for len(lines)+len(padding) < effectiveHeight {
			padding = append(padding, "")
		}
		lines = append(lines[:fillerAt], append(padding, lines[fillerAt:]...)...)
	}

	mainView := containerStyle.
		Height(maxInt(4, effectiveHeight)).
		MaxHeight(maxInt(4, effectiveHeight)).
		Render(strings.Join(lines, "\n"))
	return mainView
}

func (m PickerModel) Selected() *notes.Entry {
	return m.selected
}

func (m PickerModel) SelectedLine() int {
	return m.selectedLine
}

func (m PickerModel) Cancelled() bool {
	return m.cancelled
}

func (m PickerModel) EditRequested() bool {
	return m.edit
}

func (m PickerModel) CreateKind() string {
	return m.createKind
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
	m.refreshPreview()
	m.clampPreviewHit()
}

func (m *PickerModel) refreshPreview() {
	if len(m.matches) == 0 {
		m.preview = notes.PreviewMatch{}
		return
	}
	m.preview = m.cachedPreview(m.matches[m.cursor].Entry)
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
	if len(m.matches) == 0 || strings.TrimSpace(m.input.Value()) == "" {
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
	available := m.effectiveHeight() - 4
	if available < 6 {
		available = 6
	}
	maxItems := available / 3
	if maxItems < 2 {
		maxItems = 2
	}
	return maxItems
}

func RunPicker(entries []notes.Entry, initialQuery string, themeName string, options Options) (*notes.Entry, int, string, bool, bool, string, error) {
	theme, err := ResolveTheme(themeName)
	if err != nil {
		return nil, 0, initialQuery, false, false, "", err
	}

	model := NewPicker(entries, initialQuery, theme, options)
	programOptions := []tea.ProgramOption{}
	if options.FullScreen {
		programOptions = append(programOptions, tea.WithAltScreen())
	}
	program := tea.NewProgram(model, programOptions...)
	result, err := program.Run()
	if err != nil {
		return nil, 0, initialQuery, false, false, "", fmt.Errorf("picker: %w", err)
	}

	finalModel, ok := result.(PickerModel)
	if !ok {
		return nil, 0, initialQuery, false, false, "", fmt.Errorf("picker returned unexpected model type")
	}

	return finalModel.Selected(), finalModel.SelectedLine(), finalModel.Query(), finalModel.Cancelled(), finalModel.EditRequested(), finalModel.CreateKind(), nil
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

func (m PickerModel) inputWidth() int {
	if m.contentWidth() <= 8 {
		return 48
	}
	return m.contentWidth() - 4
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

func clipLines(lines []string, height int) []string {
	if height <= 0 || len(lines) <= height {
		return lines
	}
	if height == 1 {
		return []string{"…"}
	}

	clipped := append([]string{}, lines[:height]...)
	clipped[height-1] = "…"
	return clipped
}

func excerptLines(lines []string, limit int) []string {
	if limit <= 0 || len(lines) <= limit {
		return lines
	}
	if limit == 1 {
		return []string{"…"}
	}

	clipped := append([]string{}, lines[:limit]...)
	clipped[limit-1] = "…"
	return clipped
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

func (m *PickerModel) advancePreviewHit(delta int) {
	if len(m.matches) == 0 {
		return
	}
	occurrences := m.currentPreviewHitCount()
	if occurrences <= 1 {
		m.previewHit = 0
		return
	}
	m.previewHit = (m.previewHit + delta + occurrences) % occurrences
}

func (m *PickerModel) clampPreviewHit() {
	if len(m.matches) == 0 {
		m.previewHit = 0
		return
	}
	occurrences := m.currentPreviewHitCount()
	if occurrences <= 0 {
		m.previewHit = 0
		return
	}
	if m.previewHit >= occurrences {
		m.previewHit = occurrences - 1
	}
	if m.previewHit < 0 {
		m.previewHit = 0
	}
}

func (m PickerModel) activePreviewHit() int {
	if len(m.matches) == 0 {
		return 0
	}
	occurrences := m.currentPreviewHitCount()
	if occurrences == 0 {
		return 0
	}
	if m.previewHit >= occurrences || m.previewHit < 0 {
		return 0
	}
	return m.previewHit
}

func (m PickerModel) currentPreviewHitCount() int {
	return previewHitCount(m.preview)
}

func previewHitCount(preview notes.PreviewMatch) int {
	if count := len(preview.Snippets); count > 0 {
		return count
	}
	return len(preview.Occurrences)
}

func (m PickerModel) resultLines(width int, rowStyle, selectedStyle, detailStyle lipgloss.Style) []string {
	if len(m.matches) == 0 {
		return []string{detailStyle.Render("No matches")}
	}

	lines := make([]string, 0, len(m.visibleMatches())*5)
	visible := m.visibleMatches()
	for i, match := range visible {
		index := i + m.offset()
		entry := match.Entry
		prefix := "  "
		if index == m.cursor {
			prefix = m.theme.SelectedMark + " "
		}

		badge := resultBadge(entry)
		labelText := truncateRunes(fmt.Sprintf("%s %s", badge, match.Label), maxInt(12, width-3))
		renderedLabel := rowStyle.Render(labelText)
		if index == m.cursor {
			renderedLabel = selectedStyle.Render(padRight(labelText, maxInt(12, width-3)))
		}
		lines = append(lines, prefix+renderedLabel)

		snippet := truncateRunes(match.Detail, maxInt(12, width-4))
		if index == m.cursor {
			preview := m.cachedPreview(entry)
			if selectedSnippet := m.inlinePreviewLine(preview, width-4); selectedSnippet != "" {
				snippet = selectedSnippet
			}
		}
		lines = append(lines, "    "+snippet)
		if index == m.cursor {
			lines = append(lines, m.contextInlineLines(entry, width)...)
		}
	}
	return lines
}

func resultBadge(entry notes.Entry) string {
	switch {
	case entry.IsTemplate():
		return "[tpl]"
	case entry.HasCmd():
		return "[cmd]"
	default:
		return "[txt]"
	}
}

func (m *PickerModel) cachedPreview(entry notes.Entry) notes.PreviewMatch {
	key := previewCacheKey(entry, m.input.Value())
	if preview, ok := m.previewCache[key]; ok {
		return preview
	}
	if len(m.previewCache) > 512 {
		m.previewCache = make(map[string]notes.PreviewMatch)
	}
	preview := notes.BuildPreview(entry, m.input.Value())
	m.previewCache[key] = preview
	return preview
}

func (m PickerModel) inlinePreviewLine(preview notes.PreviewMatch, width int) string {
	lines := previewPaneLines(preview, maxInt(12, width), 1, 0, m.theme)
	if len(lines) == 0 {
		return ""
	}
	return lines[0]
}

func (m PickerModel) contextInlineLines(entry notes.Entry, totalWidth int) []string {
	if entry.HasCmd() && !entry.HasShow() && len(m.preview.Snippets) == 0 {
		return nil
	}

	width := totalWidth - 8
	if width < 24 {
		width = 24
	}
	lines := popupPreviewLines(m.preview, width, 3, m.activePreviewHit(), m.theme)
	if len(lines) == 0 {
		return nil
	}

	out := make([]string, 0, len(lines)+1)
	out = append(out, "")
	for _, line := range lines {
		out = append(out, "      "+line)
	}
	return out
}

func previewCacheKey(entry notes.Entry, query string) string {
	return entry.SourcePath + "|" + entry.Desc + "|" + query
}
