package ui

import (
	"fmt"
	"strings"

	"aoo/internal/notes"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ActionKind string

const (
	ActionRead ActionKind = "show"
	ActionRun  ActionKind = "run"
)

type EntryAction struct {
	Kind   ActionKind
	Label  string
	Detail string
	Action *notes.Action
}

type ActionPickerModel struct {
	entry     notes.Entry
	actions   []EntryAction
	options   Options
	cursor    int
	width     int
	height    int
	selected  *EntryAction
	cancelled bool
	theme     Theme
}

func NewActionPicker(entry notes.Entry, theme Theme, options Options) ActionPickerModel {
	return ActionPickerModel{
		entry:   entry,
		actions: entryActions(entry),
		theme:   theme,
		options: options,
	}
}

func (m ActionPickerModel) Init() tea.Cmd {
	return nil
}

func (m ActionPickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if len(m.actions) == 0 {
				return m, nil
			}
			action := m.actions[m.cursor]
			m.selected = &action
			return m, tea.Quit
		case "up", "ctrl+k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "ctrl+j":
			if m.cursor < len(m.actions)-1 {
				m.cursor++
			}
			return m, nil
		}
	}

	return m, nil
}

func (m ActionPickerModel) View() string {
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.TitleFG))
	detailStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.DetailFG))
	rowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.RowFG))
	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.SelectedFG)).
		Background(lipgloss.Color(m.theme.SelectedBG))
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.HelpFG))
	dividerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.DividerFG))
	contentWidth := m.contentWidth()

	lines := []string{
		titleStyle.Render(truncateRunes(m.entry.DisplayName(), maxInt(20, contentWidth))),
		detailStyle.Render(fmt.Sprintf("%d actions", len(m.actions))),
		"",
	}

	visible := m.visibleActions()
	for i, action := range visible {
		prefix := "  "
		labelText := truncateRunes(action.Label, contentWidth-3)
		label := rowStyle.Render(labelText)
		if i+m.offset() == m.cursor {
			prefix = m.theme.SelectedMark + " "
			label = selectedStyle.Render(padRight(labelText, contentWidth-3))
		}
		lines = append(lines, prefix+label)
		lines = append(lines, m.detailLine(action.Detail, m.enterHintText(action), contentWidth, detailStyle, helpStyle))
		if i != len(visible)-1 {
			lines = append(lines, "    "+dividerStyle.Render("·"))
		}
	}

	lines = append(lines, "")
	effectiveHeight := m.effectiveHeight()
	used := len(lines) + 1
	if filler := effectiveHeight - used; filler > 0 {
		for i := 0; i < filler; i++ {
			lines = append(lines, "")
		}
	}
	lines = append(lines, titleStyle.Render(truncateRunes(m.statusLine(), contentWidth)))
	if !m.options.FocusMode {
		lines = append(lines, helpStyle.Render(truncateRunes("enter choose  esc back  ↑↓ move", contentWidth)))
	}
	lines = normalizeRenderedLines(lines, contentWidth)
	return strings.Join(lines, "\n")
}

func (m ActionPickerModel) Selected() *EntryAction {
	return m.selected
}

func (m ActionPickerModel) Cancelled() bool {
	return m.cancelled
}

func (m ActionPickerModel) contentWidth() int {
	if m.width <= 0 {
		return 80
	}
	if m.width < 20 {
		return 20
	}
	return m.width - 4
}

func (m ActionPickerModel) visibleActions() []EntryAction {
	if len(m.actions) == 0 {
		return nil
	}

	maxItems := m.maxVisibleItems()
	start := m.offset()
	end := start + maxItems
	if end > len(m.actions) {
		end = len(m.actions)
	}

	return m.actions[start:end]
}

func (m ActionPickerModel) offset() int {
	if len(m.actions) == 0 {
		return 0
	}
	window := m.maxVisibleItems()
	if m.cursor < window/2 {
		return 0
	}

	start := m.cursor - window/2
	limit := len(m.actions) - window
	if limit < 0 {
		return 0
	}
	if start > limit {
		return limit
	}
	return start
}

func (m ActionPickerModel) maxVisibleItems() int {
	available := m.effectiveHeight() - 5
	if available < 6 {
		available = 6
	}
	maxItems := (available + 1) / 3
	if maxItems < 2 {
		return 2
	}
	return maxItems
}

func (m ActionPickerModel) effectiveHeight() int {
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

func (m ActionPickerModel) statusLine() string {
	return fmt.Sprintf("%d/%d", minInt(m.cursor+1, len(m.actions)), len(m.actions))
}

func RunActionPicker(entry notes.Entry, themeName string, options Options) (*EntryAction, bool, error) {
	theme, err := ResolveTheme(themeName)
	if err != nil {
		return nil, false, err
	}

	model := NewActionPicker(entry, theme, options)
	programOptions := []tea.ProgramOption{}
	if options.FullScreen {
		programOptions = append(programOptions, tea.WithAltScreen())
	}
	program := tea.NewProgram(model, programOptions...)
	result, err := program.Run()
	if err != nil {
		return nil, false, fmt.Errorf("action picker: %w", err)
	}

	finalModel, ok := result.(ActionPickerModel)
	if !ok {
		return nil, false, fmt.Errorf("action picker returned unexpected model type")
	}

	return finalModel.Selected(), finalModel.Cancelled(), nil
}

func entryActions(entry notes.Entry) []EntryAction {
	normalized := entry.ActionsList()
	actions := make([]EntryAction, 0, len(normalized))
	for i, action := range normalized {
		actionCopy := action
		label := strings.TrimSpace(action.Desc)
		if label == "" {
			switch {
			case action.IsCmd():
				label = oneLineOrFallback(action.Cmd, fmt.Sprintf("run %d", i+1))
			case action.IsShow():
				label = "show"
			default:
				label = fmt.Sprintf("action %d", i+1)
			}
		}

		item := EntryAction{
			Label:  label,
			Detail: action.DisplayValue(),
			Action: &actionCopy,
		}
		switch {
		case action.IsShow():
			item.Kind = ActionRead
		case action.IsCmd():
			item.Kind = ActionRun
		}
		if item.Kind == ActionRead && label == "show" {
			item.Label = "show note"
		}
		if item.Kind == ActionRead && entry.IsRaw() {
			item.Label = "show " + entry.SourceBadge()
		}
		actions = append(actions, item)
	}
	return actions
}

func oneLineOrFallback(value, fallback string) string {
	text := strings.TrimSpace(value)
	if text == "" {
		return fallback
	}
	return truncateRunes(strings.Join(strings.Fields(text), " "), 90)
}

func (m ActionPickerModel) enterHintText(action EntryAction) string {
	switch action.Kind {
	case ActionRead:
		if strings.HasPrefix(strings.TrimSpace(action.Label), "show ") {
			return "enter: print " + strings.TrimSpace(strings.TrimPrefix(action.Label, "show "))
		}
		return "enter: print note"
	case ActionRun:
		return "enter: run command"
	}
	return ""
}

func (m ActionPickerModel) detailLine(detail, hint string, width int, detailStyle, hintStyle lipgloss.Style) string {
	contentWidth := maxInt(12, width-4)
	detail = strings.Join(strings.Fields(strings.TrimSpace(detail)), " ")
	hint = strings.TrimSpace(hint)
	if hint == "" {
		return "    " + detailStyle.Render(truncateRunes(detail, contentWidth))
	}

	hintWidth := len([]rune(hint))
	detailWidth := contentWidth - hintWidth - 2
	if detailWidth < 1 {
		detailWidth = 1
	}
	left := truncateRunes(detail, detailWidth)
	padding := contentWidth - len([]rune(left)) - hintWidth
	if padding < 1 {
		padding = 1
	}
	if len([]rune(left))+padding+hintWidth > contentWidth {
		hint = truncateRunes(hint, maxInt(1, contentWidth-len([]rune(left))-1))
		hintWidth = len([]rune(hint))
		padding = contentWidth - len([]rune(left)) - hintWidth
		if padding < 1 {
			padding = 1
		}
	}
	return "    " + detailStyle.Render(left) + strings.Repeat(" ", padding) + hintStyle.Render(hint)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
