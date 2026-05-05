package ui

import (
	"fmt"
	"strings"
)

type Theme struct {
	Name         string
	TitleFG      string
	TitleDimFG   string
	SelectedFG   string
	SelectedBG   string
	DetailFG     string
	HelpFG       string
	InputFG      string
	InputBG      string
	InputBorder  string
	InputPrompt  string
	RowFG        string
	DividerFG    string
	SelectedMark string
	StatusOKFG   string
	StatusWarnFG string
	StatusErrFG  string
	StatusRunFG  string
}

func ResolveTheme(name string) (Theme, error) {
	key := strings.TrimSpace(strings.ToLower(name))
	if key == "" || key == "auto" {
		key = "fzf-dark"
	}

	theme, ok := themes[key]
	if !ok {
		return Theme{}, fmt.Errorf("unknown theme: %s", name)
	}

	return theme, nil
}

func ThemeNames() []string {
	return []string{
		"fzf-dark",
		"catppuccin-mocha",
		"catppuccin-latte",
		"dracula",
		"nord",
		"solarized-dark",
		"solarized-light",
	}
}

var themes = map[string]Theme{
	"fzf-dark": {
		Name:         "fzf-dark",
		TitleFG:      "#cfcfcf",
		TitleDimFG:   "#6c6c6c",
		SelectedFG:   "#ffffff",
		SelectedBG:   "#3f3f3f",
		DetailFG:     "#9a9a9a",
		HelpFG:       "#8f885e",
		InputFG:      "#f5f5f5",
		InputBG:      "#0b0b0b",
		InputBorder:  "#6c6c6c",
		InputPrompt:  "#ff5faf",
		RowFG:        "#c8c8c8",
		DividerFG:    "#3a3a3a",
		SelectedMark: ">",
		StatusOKFG:   "#87d787",
		StatusWarnFG: "#e5c07b",
		StatusErrFG:  "#ff6c6b",
		StatusRunFG:  "#61afef",
	},
	"catppuccin-mocha": {
		Name:         "catppuccin-mocha",
		TitleFG:      "#cdd6f4",
		TitleDimFG:   "#7f849c",
		SelectedFG:   "#cdd6f4",
		SelectedBG:   "#1e2336",
		DetailFG:     "#9399b2",
		HelpFG:       "#6c7086",
		InputFG:      "#cdd6f4",
		InputBG:      "#181825",
		InputBorder:  "#45475a",
		InputPrompt:  "#89b4fa",
		RowFG:        "#bac2de",
		DividerFG:    "#313244",
		SelectedMark: "›",
		StatusOKFG:   "#a6e3a1",
		StatusWarnFG: "#f9e2af",
		StatusErrFG:  "#f38ba8",
		StatusRunFG:  "#89b4fa",
	},
	"catppuccin-latte": {
		Name:         "catppuccin-latte",
		TitleFG:      "#4c4f69",
		TitleDimFG:   "#8c8fa1",
		SelectedFG:   "#4c4f69",
		SelectedBG:   "#dce0e8",
		DetailFG:     "#6c6f85",
		HelpFG:       "#8c8fa1",
		InputFG:      "#4c4f69",
		InputBG:      "#eff1f5",
		InputBorder:  "#ccd0da",
		InputPrompt:  "#1e66f5",
		RowFG:        "#5c5f77",
		DividerFG:    "#dce0e8",
		SelectedMark: "›",
		StatusOKFG:   "#40a02b",
		StatusWarnFG: "#df8e1d",
		StatusErrFG:  "#d20f39",
		StatusRunFG:  "#1e66f5",
	},
	"dracula": {
		Name:         "dracula",
		TitleFG:      "#f8f8f2",
		TitleDimFG:   "#6272a4",
		SelectedFG:   "#f8f8f2",
		SelectedBG:   "#2f3241",
		DetailFG:     "#bd93f9",
		HelpFG:       "#6272a4",
		InputFG:      "#f8f8f2",
		InputBG:      "#282a36",
		InputBorder:  "#6272a4",
		InputPrompt:  "#8be9fd",
		RowFG:        "#e9e9f4",
		DividerFG:    "#44475a",
		SelectedMark: "›",
		StatusOKFG:   "#50fa7b",
		StatusWarnFG: "#f1fa8c",
		StatusErrFG:  "#ff5555",
		StatusRunFG:  "#8be9fd",
	},
	"nord": {
		Name:         "nord",
		TitleFG:      "#eceff4",
		TitleDimFG:   "#81a1c1",
		SelectedFG:   "#eceff4",
		SelectedBG:   "#2f3b52",
		DetailFG:     "#d8dee9",
		HelpFG:       "#81a1c1",
		InputFG:      "#eceff4",
		InputBG:      "#2e3440",
		InputBorder:  "#4c566a",
		InputPrompt:  "#88c0d0",
		RowFG:        "#e5e9f0",
		DividerFG:    "#3b4252",
		SelectedMark: "›",
		StatusOKFG:   "#a3be8c",
		StatusWarnFG: "#ebcb8b",
		StatusErrFG:  "#bf616a",
		StatusRunFG:  "#88c0d0",
	},
	"solarized-dark": {
		Name:         "solarized-dark",
		TitleFG:      "#93a1a1",
		TitleDimFG:   "#586e75",
		SelectedFG:   "#93a1a1",
		SelectedBG:   "#113947",
		DetailFG:     "#839496",
		HelpFG:       "#586e75",
		InputFG:      "#93a1a1",
		InputBG:      "#073642",
		InputBorder:  "#586e75",
		InputPrompt:  "#268bd2",
		RowFG:        "#93a1a1",
		DividerFG:    "#0f414f",
		SelectedMark: "›",
		StatusOKFG:   "#859900",
		StatusWarnFG: "#b58900",
		StatusErrFG:  "#dc322f",
		StatusRunFG:  "#268bd2",
	},
	"solarized-light": {
		Name:         "solarized-light",
		TitleFG:      "#586e75",
		TitleDimFG:   "#93a1a1",
		SelectedFG:   "#586e75",
		SelectedBG:   "#eee8d5",
		DetailFG:     "#657b83",
		HelpFG:       "#93a1a1",
		InputFG:      "#586e75",
		InputBG:      "#fdf6e3",
		InputBorder:  "#eee8d5",
		InputPrompt:  "#268bd2",
		RowFG:        "#586e75",
		DividerFG:    "#eee8d5",
		SelectedMark: "›",
		StatusOKFG:   "#859900",
		StatusWarnFG: "#b58900",
		StatusErrFG:  "#dc322f",
		StatusRunFG:  "#268bd2",
	},
}
