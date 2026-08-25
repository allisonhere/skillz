package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// A small, restrained palette. Adaptive colors give a reasonable look on
// both light and dark terminal backgrounds; termenv (underneath lipgloss)
// already degrades this automatically for NO_COLOR, 256-color, and
// non-color terminals.
var (
	colorAccent   = lipgloss.AdaptiveColor{Light: "#0057B7", Dark: "#7DAFFF"}
	colorGood     = lipgloss.AdaptiveColor{Light: "#0A7A3D", Dark: "#5FD98A"}
	colorWarn     = lipgloss.AdaptiveColor{Light: "#8A5A00", Dark: "#E6B450"}
	colorMuted    = lipgloss.AdaptiveColor{Light: "#6B6B6B", Dark: "#8A8A8A"}
	colorFaint    = lipgloss.AdaptiveColor{Light: "#A0A0A0", Dark: "#555555"}
	colorFg       = lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#E6E6E6"}
	colorSelectBg = lipgloss.AdaptiveColor{Light: "#DCE8FF", Dark: "#1E2A44"}
)

var (
	styleTitle = lipgloss.NewStyle().Bold(true).Foreground(colorFg)
	styleDim   = lipgloss.NewStyle().Foreground(colorMuted)
	styleFaint = lipgloss.NewStyle().Foreground(colorFaint).Italic(true)
	styleGood  = lipgloss.NewStyle().Foreground(colorGood)
	styleWarn  = lipgloss.NewStyle().Foreground(colorWarn)
	styleAccent = lipgloss.NewStyle().Foreground(colorAccent)
	styleName   = lipgloss.NewStyle().Bold(true).Foreground(colorFg)

	styleRow = lipgloss.NewStyle().PaddingLeft(2)
	styleRowSelected = lipgloss.NewStyle().
				PaddingLeft(1).
				Background(colorSelectBg).
				Foreground(colorFg)
	styleRail = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)

	stylePanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(1, 2)

	styleFooter = lipgloss.NewStyle().Foreground(colorFaint)
	styleErr    = lipgloss.NewStyle().Foreground(colorWarn).Bold(true)
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
