package notesrepo

import (
	"bytes"
	"fmt"
	"io"
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
	if err := pushWithRebaseRetry(repoRoot, stdout, stderr); err != nil {
		return err
	}
	return nil
}

func pushWithRebaseRetry(repoRoot string, stdout, stderr io.Writer) error {
	if err := gitRun(repoRoot, stdout, stderr, "push"); err == nil {
		return nil
	}

	branch := strings.TrimSpace(gitOutput(repoRoot, "branch", "--show-current"))
	if branch == "" {
		return fmt.Errorf("git push: cannot detect current branch")
	}

	fmt.Fprintln(stdout, "[notes] remote changed, trying git pull --rebase")
	pullOutput, pullErr := gitCombined(repoRoot, "pull", "--rebase", "--autostash", "origin", branch)
	if len(pullOutput) > 0 {
		_, _ = stdout.Write(pullOutput)
	}
	if pullErr != nil {
		_ = gitRun(repoRoot, io.Discard, io.Discard, "rebase", "--abort")
		if hint := gitAuthHint(pullOutput, pullErr); hint != "" {
			return fmt.Errorf("git pull --rebase: %s", hint)
		}
		return fmt.Errorf("git pull --rebase failed, resolve notes repo conflict manually: %w", pullErr)
	}

	if err := gitRun(repoRoot, stdout, stderr, "push"); err != nil {
		return authAwareGitError("git push after rebase", nil, err)
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
	cmd := gitCommand(dir, "diff", "--cached", "--quiet", "--", relPath)
	return cmd.Run() != nil
}

func gitRun(dir string, stdout, stderr io.Writer, args ...string) error {
	cmd := gitCommand(dir, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return authAwareGitError("git "+strings.Join(args, " "), nil, err)
	}
	return nil
}

func gitCombined(dir string, args ...string) ([]byte, error) {
	cmd := gitCommand(dir, args...)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	err := cmd.Run()
	return output.Bytes(), err
}
