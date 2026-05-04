package notesrepo

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestAutoCommitPushPathCommitsAndPushes(t *testing.T) {
	remote := filepath.Join(t.TempDir(), "remote.git")
	worktree := filepath.Join(t.TempDir(), "notes")

	runGit(t, "", "init", "--bare", remote)
	runGit(t, "", "clone", remote, worktree)
	runGit(t, worktree, "config", "user.name", "aoo-test")
	runGit(t, worktree, "config", "user.email", "aoo-test@example.com")

	notePath := filepath.Join(worktree, "router.yaml")
	if err := os.WriteFile(notePath, []byte("desc: router\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if err := AutoCommitPushPath(notePath, "aoo: update router.yaml", &stdout, &stderr); err != nil {
		t.Fatal(err)
	}

	log := strings.TrimSpace(gitOutput(worktree, "log", "--format=%s", "-1"))
	if log != "aoo: update router.yaml" {
		t.Fatalf("unexpected commit message: %q", log)
	}

	remoteLog := strings.TrimSpace(runGitOutput(t, "", "--git-dir", remote, "log", "--format=%s", "-1"))
	if remoteLog != "aoo: update router.yaml" {
		t.Fatalf("expected pushed commit, got %q", remoteLog)
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmdArgs := args
	if dir != "" {
		cmdArgs = append([]string{"-C", dir}, args...)
	}

	cmd := exec.Command("git", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}
}

func runGitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmdArgs := args
	if dir != "" {
		cmdArgs = append([]string{"-C", dir}, args...)
	}

	cmd := exec.Command("git", cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, string(output))
	}
	return string(output)
}
