package app

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"aoo/internal/config"
	"aoo/internal/notes"
	"aoo/internal/notesrepo"
	"aoo/internal/templatecmd"
	"aoo/internal/ui"
)

const version = "0.1.0"

func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if len(args) > 0 {
		switch args[0] {
		case "validate":
			return runValidate(args[1:], stdout, stderr)
		case "themes":
			return runThemes(stdout)
		case "config":
			return runConfig(args[1:], stdout, stderr)
		case "set-source":
			return runSetSource(stdin, stdout, stderr)
		case "set-folder":
			return runSetFolder(args[1:], stdout, stderr)
		case "set-app-dir":
			return runSetAppDir(args[1:], stdout, stderr)
		case "set-theme":
			return runSetTheme(args[1:], stdout, stderr)
		case "version", "--version", "-v":
			_, err := fmt.Fprintf(stdout, "aoo %s\n", version)
			return err
		case "help", "--help", "-h":
			printUsage(stdout)
			return nil
		}
	}

	return runInteractive(args, stdin, stdout, stderr)
}

func runInteractive(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("aoo", flag.ContinueOnError)
	fs.SetOutput(stderr)

	dir := fs.String("dir", "", "directory with YAML notes")
	query := fs.String("query", "", "initial search query")
	themeFlag := fs.String("theme", "", "theme name")
	strict := fs.Bool("strict", false, "fail when any note file has validation errors")

	if err := fs.Parse(args); err != nil {
		return err
	}

	root, _, err := config.ResolveNotesDir(*dir)
	if err != nil {
		if _, ok := err.(config.SetupRequiredError); ok && strings.TrimSpace(*dir) == "" {
			root, err = notesrepo.BootstrapIfNeeded(stdin, stdout, stderr)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	if status, statusErr := notesrepo.CheckStatus(root); statusErr == nil {
		notesrepo.PrintHint(status, stdout)
	}

	result := notes.LoadDir(root)
	if bundled := loadBundledNotes(stderr); len(bundled.Entries) > 0 || len(bundled.Errors) > 0 {
		for _, loadErr := range bundled.Errors {
			fmt.Fprintf(stderr, "warn: %v\n", loadErr)
		}
		result.Entries = mergeEntries(result.Entries, bundled.Entries)
	}

	if len(result.Errors) > 0 {
		for _, loadErr := range result.Errors {
			fmt.Fprintf(stderr, "warn: %v\n", loadErr)
		}
		if *strict {
			return errors.New("validation errors found")
		}
	}

	if len(result.Entries) == 0 {
		return fmt.Errorf("no notes found in %s", root)
	}

	themeName, _, err := config.ResolveTheme(*themeFlag)
	if err != nil {
		return err
	}

	selected, cancelled, err := ui.RunPicker(result.Entries, *query, themeName)
	if err != nil || cancelled || selected == nil {
		return err
	}

	if selected.IsTemplate() {
		return runTemplate(*selected, stdin, stdout, stderr)
	}

	if selected.IsRun() {
		return runCommand(*selected, stdout, stderr)
	}

	printNote(*selected, stdout)
	return nil
}

func runTemplate(entry notes.Entry, stdin io.Reader, stdout, stderr io.Writer) error {
	values, err := builtInTemplateValues()
	if err != nil {
		return err
	}

	prepared, confirmed, err := templatecmd.Prompt(entry, stdin, stdout, values)
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Fprintln(stdout, "cancelled")
		return nil
	}

	if banner := strings.TrimSpace(entry.Banner); banner != "" {
		fmt.Fprintln(stdout, renderBanner(entry.Desc, banner))
	}

	cmd := exec.Command("/bin/sh", "-lc", prepared.Command)
	cmd.Stdin = os.Stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

func runValidate(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	fs.SetOutput(stderr)
	dir := fs.String("dir", "", "directory with YAML notes")
	if err := fs.Parse(args); err != nil {
		return err
	}

	root, _, err := config.ResolveNotesDir(*dir)
	if err != nil {
		return err
	}

	result := notes.LoadDir(root)
	if len(result.Errors) == 0 {
		_, err = fmt.Fprintf(stdout, "OK: %d entries loaded from %s\n", len(result.Entries), root)
		return err
	}

	for _, loadErr := range result.Errors {
		fmt.Fprintf(stderr, "ERROR: %v\n", loadErr)
	}
	return fmt.Errorf("validation failed: %d file(s) with errors", len(result.Errors))
}

func loadBundledNotes(stderr io.Writer) notes.LoadResult {
	appDir, _, err := config.ResolveAppDir("")
	if err != nil || strings.TrimSpace(appDir) == "" {
		return notes.LoadResult{}
	}

	examplesDir := filepath.Join(appDir, "examples", "notes")
	if _, statErr := os.Stat(examplesDir); statErr != nil {
		return notes.LoadResult{}
	}

	return notes.LoadDir(examplesDir)
}

func runCommand(entry notes.Entry, stdout, stderr io.Writer) error {
	if banner := strings.TrimSpace(entry.Banner); banner != "" {
		fmt.Fprintln(stdout, renderBanner(entry.Desc, banner))
	}

	cmd := exec.Command("/bin/sh", "-lc", entry.Run)
	cmd.Stdin = os.Stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

func runConfig(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		printConfigUsage(stdout)
		return nil
	}

	switch args[0] {
	case "show":
		return runConfigShow(stdout)
	case "set-source":
		return runSetSource(os.Stdin, stdout, stderr)
	case "set-folder":
		return runSetFolder(args[1:], stdout, stderr)
	case "set-app-dir":
		return runSetAppDir(args[1:], stdout, stderr)
	case "set-theme":
		return runSetTheme(args[1:], stdout, stderr)
	default:
		return fmt.Errorf("unknown config subcommand: %s", args[0])
	}
}

func runSetSource(stdin io.Reader, stdout, stderr io.Writer) error {
	_, err := notesrepo.SetupSource(stdin, stdout, stderr)
	return err
}

func runSetFolder(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("set-folder", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() != 1 {
		return errors.New("usage: aoo set-folder /path/to/notes")
	}

	path, err := config.SetNotesDir(fs.Arg(0))
	if err != nil {
		return err
	}

	configPath, err := config.ConfigPath()
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "configured notes_dir: %s\nconfig file: %s\n", path, configPath)
	return err
}

func runConfigShow(stdout io.Writer) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	configPath, err := config.ConfigPath()
	if err != nil {
		return err
	}

	root, source, resolveErr := config.ResolveNotesDir("")
	if resolveErr != nil {
		if _, ok := resolveErr.(config.SetupRequiredError); ok {
			root = ""
			source = "not configured"
		} else {
			return resolveErr
		}
	}

	themeName, themeSource, themeErr := config.ResolveTheme("")
	if themeErr != nil {
		return themeErr
	}
	appDir, appSource, appErr := config.ResolveAppDir("")
	if appErr != nil {
		return appErr
	}

	fmt.Fprintf(stdout, "config file: %s\n", configPath)
	fmt.Fprintf(stdout, "notes_dir: %s\n", emptyIfUnset(cfg.NotesDir))
	fmt.Fprintf(stdout, "notes_repo: %s\n", emptyIfUnset(cfg.NotesRepo))
	fmt.Fprintf(stdout, "active dir: %s\n", emptyIfUnset(root))
	fmt.Fprintf(stdout, "active source: %s\n", source)
	fmt.Fprintf(stdout, "app_dir: %s\n", emptyIfUnset(cfg.AppDir))
	fmt.Fprintf(stdout, "active app dir: %s\n", emptyIfUnset(appDir))
	fmt.Fprintf(stdout, "app dir source: %s\n", emptyIfUnset(appSource))
	fmt.Fprintf(stdout, "theme: %s\n", emptyIfUnset(cfg.Theme))
	fmt.Fprintf(stdout, "active theme: %s\n", themeName)
	fmt.Fprintf(stdout, "theme source: %s\n", themeSource)
	return nil
}

func runSetAppDir(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("set-app-dir", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() != 1 {
		return errors.New("usage: aoo set-app-dir /path/to/aoo")
	}

	path, err := config.SetAppDir(fs.Arg(0))
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "configured app_dir: %s\n", path)
	return err
}

func runSetTheme(args []string, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("set-theme", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() != 1 {
		return errors.New("usage: aoo set-theme THEME")
	}

	themeName := fs.Arg(0)
	if _, err := ui.ResolveTheme(themeName); err != nil {
		return err
	}

	saved, err := config.SetTheme(themeName)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(stdout, "configured theme: %s\n", saved)
	return err
}

func runThemes(stdout io.Writer) error {
	fmt.Fprintln(stdout, "Themes:")
	for _, name := range ui.ThemeNames() {
		fmt.Fprintf(stdout, "  %s\n", name)
	}
	return nil
}

func printNote(entry notes.Entry, stdout io.Writer) {
	fmt.Fprintln(stdout, entry.Desc)
	fmt.Fprintln(stdout, strings.TrimRight(entry.Note, "\n"))
}

func renderBanner(title, message string) string {
	lines := strings.Split(strings.TrimSpace(message), "\n")
	width := len(title)
	for _, line := range lines {
		if len(line) > width {
			width = len(line)
		}
	}
	width += 4

	var b strings.Builder
	border := "+" + strings.Repeat("-", width+2) + "+"
	b.WriteString(border + "\n")
	b.WriteString(fmt.Sprintf("| %-*s |\n", width, title))
	b.WriteString("|" + strings.Repeat("-", width+2) + "|\n")
	for _, line := range lines {
		b.WriteString(fmt.Sprintf("| %-*s |\n", width, line))
	}
	b.WriteString(border + "\n")
	return b.String()
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "aoo")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  aoo                 open notes")
	fmt.Fprintln(w, "  aoo validate        validate yaml notes")
	fmt.Fprintln(w, "  aoo set-source      notes source wizard")
	fmt.Fprintln(w, "  aoo set-folder DIR  set notes folder")
	fmt.Fprintln(w, "  aoo set-app-dir DIR set aoo repo folder")
	fmt.Fprintln(w, "  aoo set-theme NAME  set theme")
	fmt.Fprintln(w, "  aoo themes          list themes")
	fmt.Fprintln(w, "  aoo config show     show current config")
	fmt.Fprintln(w, "  aoo version")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "Examples:")
	fmt.Fprintln(w, "  aoo")
	fmt.Fprintln(w, "  aoo --query chash")
	fmt.Fprintln(w, "  aoo --theme catppuccin-latte")
	fmt.Fprintln(w, "  aoo --query nmap")
	fmt.Fprintln(w, "  aoo set-source")
	fmt.Fprintln(w, "  aoo set-folder ~/notes")
	fmt.Fprintln(w, "  aoo set-app-dir ~/workspace/aoo")
	fmt.Fprintln(w, "  aoo set-theme nord")
}

func printConfigUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  aoo config show")
	fmt.Fprintln(w, "  aoo config set-folder PATH")
	fmt.Fprintln(w, "  aoo config set-app-dir PATH")
	fmt.Fprintln(w, "  aoo config set-theme THEME")
}

func emptyIfUnset(value string) string {
	if strings.TrimSpace(value) == "" {
		return "(not set)"
	}
	return value
}

func builtInTemplateValues() (map[string]string, error) {
	values := map[string]string{}

	notesDir, _, err := config.ResolveNotesDir("")
	if err == nil && strings.TrimSpace(notesDir) != "" {
		values["aoo_notes_dir"] = notesDir
	}

	appDir, _, err := config.ResolveAppDir("")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(appDir) != "" {
		values["aoo_app_dir"] = appDir
	}

	return values, nil
}

func mergeEntries(primary, secondary []notes.Entry) []notes.Entry {
	merged := make([]notes.Entry, 0, len(primary)+len(secondary))
	seen := map[string]struct{}{}

	appendEntry := func(entry notes.Entry) {
		key := entry.Desc + "|" + entry.Action() + "|" + entry.DisplayValue()
		if _, exists := seen[key]; exists {
			return
		}
		seen[key] = struct{}{}
		merged = append(merged, entry)
	}

	for _, entry := range primary {
		appendEntry(entry)
	}
	for _, entry := range secondary {
		appendEntry(entry)
	}

	return merged
}
