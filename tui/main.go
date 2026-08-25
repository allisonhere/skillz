// Command skillz-tui is an interactive picker for installing and
// uninstalling the skills in this repo across the local agents that support
// them. It's the TUI front end for what install-skill does from a shell;
// the two share the same target definitions and never disagree on status.
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	repoDir, err := findRepoRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, "skillz-tui:", err)
		os.Exit(1)
	}

	skills, err := discoverSkills(repoDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "skillz-tui:", err)
		os.Exit(1)
	}
	if len(skills) == 0 {
		fmt.Fprintf(os.Stderr, "skillz-tui: no skills found in %s (looked for */SKILL.md)\n", repoDir)
		os.Exit(1)
	}

	m := newModel(repoDir, skills, loadTargets())

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "skillz-tui:", err)
		os.Exit(1)
	}
}

// findRepoRoot locates the skillz repo: $SKILLZ_REPO if set, otherwise the
// nearest ancestor of this binary (or, failing that, of the current
// directory) that contains both install-skill and check-skill.
func findRepoRoot() (string, error) {
	if v := os.Getenv("SKILLZ_REPO"); v != "" {
		if looksLikeRepo(v) {
			return v, nil
		}
		return "", fmt.Errorf("SKILLZ_REPO=%s doesn't look like the skillz repo (no install-skill/check-skill)", v)
	}

	if exe, err := os.Executable(); err == nil {
		if real, err := filepath.EvalSymlinks(exe); err == nil {
			if root, ok := searchUpward(filepath.Dir(real)); ok {
				return root, nil
			}
		}
	}

	if cwd, err := os.Getwd(); err == nil {
		if root, ok := searchUpward(cwd); ok {
			return root, nil
		}
	}

	return "", errors.New("couldn't find the skillz repo; set SKILLZ_REPO=/path/to/skillz")
}

func searchUpward(start string) (string, bool) {
	dir := start
	for i := 0; i < 6; i++ {
		if looksLikeRepo(dir) {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", false
}

func looksLikeRepo(dir string) bool {
	_, err1 := os.Stat(filepath.Join(dir, "install-skill"))
	_, err2 := os.Stat(filepath.Join(dir, "check-skill"))
	return err1 == nil && err2 == nil
}
