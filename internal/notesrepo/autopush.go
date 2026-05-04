package notesrepo

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
)

func AutoCommitPushPath(path, message string, stdout, stderr io.Writer) error {
	path = strings.TrimSpace(path)
	message = strings.TrimSpace(message)
	if path == "" || message == "" {
		return nil
	}

	repoRoot := strings.TrimSpace(repoRootForPath(path))
	if repoRoot == "" {
		return nil
	}

	relPath, err := filepath.Rel(repoRoot, path)
	if err != nil {
		return fmt.Errorf("resolve repo-relative path: %w", err)
	}

	if err := gitRun(repoRoot, stdout, stderr, "add", "--", relPath); err != nil {
		return fmt.Errorf("git add %s: %w", relPath, err)
	}
	if !hasStagedChanges(repoRoot, relPath) {
		return nil
	}

	if err := gitRun(repoRoot, stdout, stderr, "commit", "-m", message, "--", relPath); err != nil {
		return fmt.Errorf("git commit %s: %w", relPath, err)
	}

	if strings.TrimSpace(gitOutput(repoRoot, "remote", "get-url", "origin")) == "" {
		return nil
	}
	if err := gitRun(repoRoot, stdout, stderr, "push"); err != nil {
		return fmt.Errorf("git push: %w", err)
	}
	return nil
}

func repoRootForPath(path string) string {
	dir := path
	info, err := filepath.Abs(path)
	if err == nil {
		dir = info
	}
	dir = filepath.Dir(dir)
	return strings.TrimSpace(gitOutput(dir, "rev-parse", "--show-toplevel"))
}

func hasStagedChanges(dir, relPath string) bool {
	cmd := exec.Command("git", "-C", dir, "diff", "--cached", "--quiet", "--", relPath)
	return cmd.Run() != nil
}

func gitRun(dir string, stdout, stderr io.Writer, args ...string) error {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}
