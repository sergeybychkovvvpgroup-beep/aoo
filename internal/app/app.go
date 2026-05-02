package app

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/sergeyb/aoo/internal/config"
	"github.com/sergeyb/aoo/internal/notes"
	"github.com/sergeyb/aoo/internal/templatecmd"
	"github.com/sergeyb/aoo/internal/ui"
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
		case "set-folder":
			return runSetFolder(args[1:], stdout, stderr)
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
		return err
	}

	result := notes.LoadDir(root)
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
	prepared, confirmed, err := templatecmd.Prompt(entry, stdin, stdout)
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

func runCommand(entry notes.Entry, stdout, stderr io.Writer) error {
	if err := runPreCheck(entry, stdout, stderr); err != nil {
		return err
	}

	if banner := strings.TrimSpace(entry.Banner); banner != "" {
		fmt.Fprintln(stdout, renderBanner(entry.Desc, banner))
	}

	cmd := exec.Command("/bin/sh", "-lc", entry.Run)
	cmd.Stdin = os.Stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

func runPreCheck(entry notes.Entry, stdout, stderr io.Writer) error {
	check := strings.TrimSpace(entry.Check)
	if check == "" {
		return nil
	}

	fmt.Fprintf(stdout, "[check] %s\n", entry.Desc)

	cmd := exec.Command("/bin/sh", "-lc", check)
	cmd.Stdin = nil
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(entry.CheckError)
		if msg == "" {
			msg = "pre-check failed"
		}
		return fmt.Errorf("%s\ncheck: %s", msg, check)
	}

	return nil
}

func runConfig(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		printConfigUsage(stdout)
		return nil
	}

	switch args[0] {
	case "show":
		return runConfigShow(stdout)
	case "set-folder":
		return runSetFolder(args[1:], stdout, stderr)
	case "set-theme":
		return runSetTheme(args[1:], stdout, stderr)
	default:
		return fmt.Errorf("unknown config subcommand: %s", args[0])
	}
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

	fmt.Fprintf(stdout, "config file: %s\n", configPath)
	fmt.Fprintf(stdout, "notes_dir: %s\n", emptyIfUnset(cfg.NotesDir))
	fmt.Fprintf(stdout, "active dir: %s\n", emptyIfUnset(root))
	fmt.Fprintf(stdout, "active source: %s\n", source)
	fmt.Fprintf(stdout, "theme: %s\n", emptyIfUnset(cfg.Theme))
	fmt.Fprintf(stdout, "active theme: %s\n", themeName)
	fmt.Fprintf(stdout, "theme source: %s\n", themeSource)
	return nil
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
	fmt.Fprintf(stdout, "- desc: %s\n", entry.Desc)
	fmt.Fprintln(stdout, "  note: |")
	for _, line := range strings.Split(strings.TrimRight(entry.Note, "\n"), "\n") {
		fmt.Fprintf(stdout, "    %s\n", line)
	}
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
	fmt.Fprintln(w, "  aoo set-folder DIR  set notes folder")
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
	fmt.Fprintln(w, "  aoo set-folder ~/notes")
	fmt.Fprintln(w, "  aoo set-theme nord")
}

func printConfigUsage(w io.Writer) {
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  aoo config show")
	fmt.Fprintln(w, "  aoo config set-folder PATH")
	fmt.Fprintln(w, "  aoo config set-theme THEME")
}

func emptyIfUnset(value string) string {
	if strings.TrimSpace(value) == "" {
		return "(not set)"
	}
	return value
}
