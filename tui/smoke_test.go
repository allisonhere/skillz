package main

import (
	"bytes"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/teatest"
)

func TestSmokeListAndManageAndQuit(t *testing.T) {
	repoDir := "testdata_repo"
	skills := []Skill{{Name: "modern-tui", Description: "A test skill", Dir: "/nonexistent/modern-tui"}}
	targets := []Target{{Name: "agents", Base: "/nonexistent/.agents/skills"}}

	m := newModel(repoDir, skills, targets)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(100, 30))

	// List screen should show the skill name.
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("modern-tui")) && bytes.Contains(b, []byte("skillz"))
	}, teatest.WithDuration(2*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return bytes.Contains(b, []byte("space toggle")) || bytes.Contains(b, []byte("agent not found"))
	}, teatest.WithDuration(2*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyEsc})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})

	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
}

// renderNeverExceedsWidth is the bounded-render invariant from
// modern-tui/references/testing.md: at any terminal size, no emitted line
// may exceed the requested width.
func TestRenderNeverExceedsWidth(t *testing.T) {
	skills := []Skill{
		{Name: "modern-tui", Description: "Design and implement polished, keyboard-first terminal apps", Dir: "/x/modern-tui"},
		{Name: "a-very-long-skill-name-that-should-truncate-cleanly", Description: "こんにちは 🎉 wide and emoji content to stress the width math", Dir: "/x/long"},
	}
	targets := []Target{
		{Name: "agents", Base: "/nonexistent/.agents/skills"},
		{Name: "codex", Base: "/nonexistent/.codex/skills"},
		{Name: "claude", Base: "/nonexistent/.claude/skills"},
		{Name: "hermes", Base: "/nonexistent/.hermes/skills", Copy: true},
	}

	for w := 20; w <= 140; w += 7 {
		for h := 5; h <= 40; h += 5 {
			m := newModel("testdata_repo", skills, targets)
			m.width, m.height, m.ready = w, h, true
			for _, line := range strings.Split(m.View(), "\n") {
				if got := lipgloss.Width(line); got > w {
					t.Fatalf("list w=%d h=%d: line exceeded width: got=%d line=%q", w, h, got, line)
				}
			}
			m.screen = screenManage
			for _, line := range strings.Split(m.View(), "\n") {
				if got := lipgloss.Width(line); got > w {
					t.Fatalf("manage w=%d h=%d: line exceeded width: got=%d line=%q", w, h, got, line)
				}
			}
		}
	}
}
