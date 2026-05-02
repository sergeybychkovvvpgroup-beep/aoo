package notesrepo

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"aoo/internal/config"
)

type Status struct {
	IsRepo     bool
	RemoteURL  string
	Dirty      bool
	DirtyFiles int
	Ahead      int
	Behind     int
}

func DefaultNotesDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "aoo", "notes"), nil
}

func BootstrapIfNeeded(stdin io.Reader, stdout, stderr io.Writer) (string, error) {
	root, _, err := config.ResolveNotesDir("")
	if err == nil && strings.TrimSpace(root) != "" {
		return root, nil
	}

	if _, ok := err.(config.SetupRequiredError); !ok {
		return "", err
	}

	return SetupSource(stdin, stdout, stderr)
}

func SetupSource(stdin io.Reader, stdout, stderr io.Writer) (string, error) {
	reader := bufio.NewReader(stdin)
	defaultDir, err := DefaultNotesDir()
	if err != nil {
		return "", err
	}

	fmt.Fprintln(stdout, "aoo notes source")
	fmt.Fprintln(stdout, "1. repo")
	fmt.Fprintln(stdout, "2. local folder")
	fmt.Fprint(stdout, "(1/2?) [1]: ")
	mode, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	mode = strings.TrimSpace(mode)
	if mode == "" {
		mode = "1"
	}

	if mode == "1" {
		fmt.Fprintf(stdout, "repo url: ")
		repoURL, readErr := reader.ReadString('\n')
		if readErr != nil && readErr != io.EOF {
			return "", readErr
		}
		repoURL = strings.TrimSpace(repoURL)
		if repoURL == "" {
			return "", fmt.Errorf("repo url is required for source type repo")
		}
		if err := cloneRepo(repoURL, defaultDir, stdout, stderr); err != nil {
			pubkeys, _ := os.ReadFile(filepath.Join(os.Getenv("HOME"), ".ssh", "id_ed25519.pub"))
			if len(pubkeys) == 0 {
				pubkeys, _ = os.ReadFile(filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa.pub"))
			}
			return "", fmt.Errorf("clone notes repo failed\nif repo is private, add this public key to Deploy Keys:\n%s\noriginal error: %w", strings.TrimSpace(string(pubkeys)), err)
		}
		if err := config.SetNotesRepo(repoURL); err != nil {
			return "", err
		}
		if _, err := config.SetNotesDir(defaultDir); err != nil {
			return "", err
		}
		fmt.Fprintf(stdout, "configured notes_dir: %s\n", defaultDir)
		return defaultDir, nil
	}

	fmt.Fprintf(stdout, "notes folder [%s]: ", defaultDir)
	value, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	value = strings.TrimSpace(value)
	if value != "" {
		defaultDir = value
	}
	defaultDir, err = filepath.Abs(defaultDir)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(defaultDir, 0o755); err != nil {
		return "", err
	}
	if _, err := config.SetNotesDir(defaultDir); err != nil {
		return "", err
	}
	fmt.Fprintf(stdout, "configured notes_dir: %s\n", defaultDir)
	return defaultDir, nil
}

func CheckStatus(dir string) (Status, error) {
	if strings.TrimSpace(dir) == "" || !isGitRepo(dir) {
		return Status{}, nil
	}

	status := Status{IsRepo: true}
	status.RemoteURL = strings.TrimSpace(gitOutput(dir, "remote", "get-url", "origin"))
	porcelain := strings.TrimSpace(gitOutput(dir, "status", "--porcelain"))
	if porcelain != "" {
		status.Dirty = true
		status.DirtyFiles = len(strings.Split(porcelain, "\n"))
	}

	status.Ahead, status.Behind = aheadBehind(dir)
	return status, nil
}

func PrintHint(status Status, stdout io.Writer) {
	if !status.IsRepo {
		return
	}

	switch {
	case status.Dirty:
		fmt.Fprintf(stdout, "[notes] %d local change(s). template: notes git add commit push\n", status.DirtyFiles)
	case status.Ahead > 0 && status.Behind > 0:
		fmt.Fprintf(stdout, "[notes] diverged: %d ahead / %d behind\n", status.Ahead, status.Behind)
	case status.Ahead > 0:
		fmt.Fprintf(stdout, "[notes] %d local commit(s) to push\n", status.Ahead)
	case status.Behind > 0:
		fmt.Fprintf(stdout, "[notes] %d fetched commit(s) available. template: notes git pull\n", status.Behind)
	}
}

func cloneRepo(repoURL, targetDir string, stdout, stderr io.Writer) error {
	if err := os.MkdirAll(filepath.Dir(targetDir), 0o755); err != nil {
		return err
	}
	cmd := exec.Command("git", "clone", repoURL, targetDir)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

func isGitRepo(dir string) bool {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--is-inside-work-tree")
	return cmd.Run() == nil
}

func git(dir string, args ...string) error {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	return cmd.Run()
}

func gitOutput(dir string, args ...string) string {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return ""
	}
	return stdout.String()
}

func aheadBehind(dir string) (ahead int, behind int) {
	out := strings.TrimSpace(gitOutput(dir, "rev-list", "--left-right", "--count", "@{upstream}...HEAD"))
	if out == "" {
		return 0, 0
	}
	fmt.Sscanf(out, "%d %d", &behind, &ahead)
	return ahead, behind
}
