package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// Deliberately NOT using lipgloss.AdaptiveColor / custom hex here. Adaptive
// colors depend on correctly detecting whether the terminal background is
// light or dark (an OSC 11 query with a fallback guess when that's
// unsupported), and a wrong guess makes half the palette pick colors tuned
// for the opposite background - unreadable, not just low-contrast. Standard
// 4-bit ANSI color numbers sidestep that: the terminal maps them to
// whatever the user's own theme considers readable against their own
// background, because that's what the rest of their shell already uses.
// Faint/Reverse are SGR attributes, not colors, so they carry no background
// assumption at all - see modern-tui/references/capabilities.md.
var (
	colorGood   = lipgloss.Color("2") // green
	colorWarn   = lipgloss.Color("3") // yellow
	colorErr    = lipgloss.Color("1") // red
	colorAccent = lipgloss.Color("6") // cyan
)

var (
	styleTitle  = lipgloss.NewStyle().Bold(true)
	styleDim    = lipgloss.NewStyle().Faint(true)
	styleFaint  = lipgloss.NewStyle().Faint(true).Italic(true)
	styleGood   = lipgloss.NewStyle().Foreground(colorGood)
	styleWarn   = lipgloss.NewStyle().Foreground(colorWarn)
	styleAccent = lipgloss.NewStyle().Foreground(colorAccent)
	styleName   = lipgloss.NewStyle().Bold(true)

	styleRow         = lipgloss.NewStyle().PaddingLeft(2)
	styleRowSelected = lipgloss.NewStyle().PaddingLeft(1).Reverse(true).Bold(true)
	styleRail        = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)

	stylePanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(1, 2)

	styleFooter = lipgloss.NewStyle().Faint(true)
	styleErr    = lipgloss.NewStyle().Foreground(colorErr).Bold(true)
)

// statusGlyph renders a target's status as a single, colored, fixed-width
// cell: a filled dot for installed, a hollow dot for not-installed, "!" for
// a real conflict, and a dim dash when the agent isn't on this machine.
func statusGlyph(s Status) string {
	switch s {
	case StatusInstalled:
		return styleGood.Render("●")
	case StatusConflict:
		return styleWarn.Render("!")
	case StatusAgentMissing:
		return styleFaint.Render("-")
	default:
		return styleDim.Render("○")
	}
}

// truncateCells shortens s to at most w display cells, ANSI-unaware (call it
// on plain text only, before styling), appending an ellipsis when it had to
// cut. See modern-tui's text-and-width reference: never byte/rune-slice a
// styled string.
func truncateCells(s string, w int) string {
	if w <= 0 {
		return ""
	}
	if runewidth.StringWidth(s) <= w {
		return s
	}
	if w == 1 {
		return "…"
	}
	return runewidth.Truncate(s, w, "…")
}

// padCells right-pads plain text to exactly w display cells, truncating
// first if needed, so table-like columns stay aligned regardless of
// wide/narrow runes.
func padCells(s string, w int) string {
	s = truncateCells(s, w)
	if gap := w - runewidth.StringWidth(s); gap > 0 {
		s += strings.Repeat(" ", gap)
	}
	return s
}
