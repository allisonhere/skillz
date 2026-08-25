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

	// Every target here resolves to StatusAgentMissing (the nonexistent
	// bases), which carries the longest status label ("agent not found")
	// - the worst case for the manage panel's row-width budget.
	wantManageLines := -1

	for w := 20; w <= 140; w += 7 {
		for h := 5; h <= 40; h += 5 {
			// Exercise both the selected-row and unselected-row style path,
			// and call the screen renderers directly (not the outer View(),
			// which runs everything through clampWidth) - clampWidth's job
			// is to be a last-resort net for genuinely pathological cases,
			// not to silently mask an off-by-one in a row's own width
			// budget. A prior bug (stale PaddingLeft double-counted with an
			// explicit prefix, making selected rows 1 cell and unselected
			// rows 2 cells too wide) passed the old version of this check
			// every time because clampWidth quietly ate the overflow.
			if w >= 44 { // nameW/descW's own floor (see viewList) is genuinely fixed-size below this
				for _, cursor := range []int{0, 1} {
					m := newModel("testdata_repo", skills, targets)
					m.width, m.height, m.ready = w, h, true
					m.cursor = cursor
					for _, line := range strings.Split(m.viewList(), "\n") {
						if got := lipgloss.Width(line); got > w {
							t.Fatalf("list w=%d h=%d cursor=%d: unclamped line exceeded width: got=%d line=%q", w, h, cursor, got, line)
						}
					}
				}
			}

			m := newModel("testdata_repo", skills, targets)
			m.width, m.height, m.ready = w, h, true
			for _, line := range strings.Split(m.View(), "\n") {
				if got := lipgloss.Width(line); got > w {
					t.Fatalf("list w=%d h=%d: line exceeded width: got=%d line=%q", w, h, got, line)
				}
			}
			m.screen = screenManage
			if w >= 52 { // panelWidth's own floor (see viewManage) is genuinely fixed-size below this
				for _, tcursor := range []int{0, 3} {
					m.tcursor = tcursor
					for _, line := range strings.Split(m.viewManage(), "\n") {
						if got := lipgloss.Width(line); got > w {
							t.Fatalf("manage w=%d h=%d tcursor=%d: unclamped line exceeded width: got=%d line=%q", w, h, tcursor, got, line)
						}
					}
				}
			}
			m.tcursor = 0
			manageOut := m.View()
			for _, line := range strings.Split(manageOut, "\n") {
				if got := lipgloss.Width(line); got > w {
					t.Fatalf("manage w=%d h=%d: line exceeded width: got=%d line=%q", w, h, got, line)
				}
			}
			// A well-formed panel never internally word-wraps a row just
			// because the destination path or status label is long - it
			// truncates instead. If it does wrap, this line count changes
			// with the row-budget arithmetic even though every individual
			// line still passes the width check above (the earlier real
			// bug: an overflowing row silently broke onto an extra bordered
			// line instead of failing loudly).
			if w >= 52 { // panelWidth (m.width-6, floored at 46) is genuinely fixed-size below this
				got := strings.Count(manageOut, "\n")
				if wantManageLines == -1 {
					wantManageLines = got
				} else if got != wantManageLines {
					t.Fatalf("manage w=%d h=%d: panel line count changed (got %d, want %d) - a row likely word-wrapped instead of truncating",
						w, h, got, wantManageLines)
				}
			}
		}
	}
}
