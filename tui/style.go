package main

import (
	"fmt"
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
//
// The one exception is titleGradient below: plain truecolor hex (not
// AdaptiveColor) used purely as decoration on the wordmark and rule. It
// never queries or guesses the background - termenv just steps it down to
// 256/16/mono based on COLORTERM/TERM/NO_COLOR like any other color here -
// so it doesn't reintroduce the bug this comment is warning about.
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

	// No padding here: the caller already prepends an explicit 2-cell
	// prefix ("  " or the "❯ " rail) before applying either style, so
	// adding more here would silently push rows past their width budget.
	styleRow         = lipgloss.NewStyle()
	styleRowSelected = lipgloss.NewStyle().Reverse(true).Bold(true)
	styleRail        = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	styleDivider     = lipgloss.NewStyle().Faint(true)

	stylePanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorAccent).
			Padding(1, 2)

	styleFooter = lipgloss.NewStyle().Faint(true)
	styleErr    = lipgloss.NewStyle().Foreground(colorErr).Bold(true)
)

// Small vocabulary of Unicode glyphs used throughout - basic block and
// geometric shapes only (modern-tui's "basic" capability level), never
// nerd-font glyphs or emoji, so this looks right on any modern terminal
// with no opt-in required.
const (
	iconMark   = "◆" // decorative bullet, e.g. before the title
	iconFocus  = "❯" // "this row/target is focused"
	meterFull  = "▰"
	meterEmpty = "▱"
)

// titleGradient is the one deliberately decorative flourish in this app: a
// truecolor sweep across the wordmark and the rule beneath it.
var titleGradient = []string{"#22D3EE", "#818CF8", "#F472B6"}

// gradientColors samples n evenly spaced colors across a multi-stop hex
// gradient.
func gradientColors(stops []string, n int) []lipgloss.Color {
	if n <= 0 {
		return nil
	}
	rgbs := make([][3]int, len(stops))
	for i, s := range stops {
		rgbs[i] = hexRGB(s)
	}
	out := make([]lipgloss.Color, n)
	segs := len(stops) - 1
	for i := 0; i < n; i++ {
		var t float64
		if n > 1 {
			t = float64(i) / float64(n-1)
		}
		segF := t * float64(segs)
		seg := int(segF)
		if seg >= segs {
			seg = segs - 1
		}
		if seg < 0 {
			seg = 0
		}
		localT := segF - float64(seg)
		c1, c2 := rgbs[seg], rgbs[seg+1]
		r := lerp(c1[0], c2[0], localT)
		g := lerp(c1[1], c2[1], localT)
		b := lerp(c1[2], c2[2], localT)
		out[i] = lipgloss.Color(fmt.Sprintf("#%02X%02X%02X", r, g, b))
	}
	return out
}

func lerp(a, b int, t float64) int {
	return a + int(float64(b-a)*t)
}

func hexRGB(hex string) [3]int {
	hex = strings.TrimPrefix(hex, "#")
	var r, g, b int
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return [3]int{r, g, b}
}

// gradientText colors s left-to-right across stops. s must be plain ASCII -
// it's only ever used on the fixed "skillz" wordmark - so per-byte
// iteration is safe and width-exact.
func gradientText(s string, stops []string, bold bool) string {
	cols := gradientColors(stops, len(s))
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		st := lipgloss.NewStyle().Foreground(cols[i])
		if bold {
			st = st.Bold(true)
		}
		b.WriteString(st.Render(string(s[i])))
	}
	return b.String()
}

// gradientRule draws a full-width horizontal rule that sweeps across stops,
// exactly w cells.
func gradientRule(w int, stops []string) string {
	if w <= 0 {
		return ""
	}
	cols := gradientColors(stops, w)
	var b strings.Builder
	for i := 0; i < w; i++ {
		b.WriteString(lipgloss.NewStyle().Foreground(cols[i]).Render("─"))
	}
	return b.String()
}

// meter renders a coverage fraction as a fixed-width segmented bar. Mirrors
// the pattern documented in modern-tui/references/graphs.md.
func meter(frac float64, width int) string {
	if width <= 0 {
		return ""
	}
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	filled := int(frac*float64(width) + 0.5)
	return styleAccent.Render(strings.Repeat(meterFull, filled)) +
		styleFaint.Render(strings.Repeat(meterEmpty, width-filled))
}

// keyHints renders a footer-style list of "key label" pairs joined by a
// faint separator, e.g. "↵ manage   ·   r refresh   ·   q quit".
func keyHints(pairs ...[2]string) string {
	parts := make([]string, len(pairs))
	for i, p := range pairs {
		parts[i] = styleAccent.Bold(true).Render(p[0]) + styleFooter.Render(" "+p[1])
	}
	return strings.Join(parts, styleFooter.Render("   ·   "))
}

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
