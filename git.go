package main

import (
	"os/exec"
	"strings"
)

// gitGet runs a git command and returns trimmed stdout, or "" on error.
func gitGet(args ...string) string {
	out, err := exec.Command("git", args...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// gitSet runs git config with the given scope, key, and value.
func gitSet(scope, key, value string) error {
	return exec.Command("git", "config", scope, key, value).Run()
}

// insideRepo reports whether the current directory is inside a Git repository.
func insideRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Stderr = nil
	return cmd.Run() == nil
}
