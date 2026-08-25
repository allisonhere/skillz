package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type screen int

const (
	screenList screen = iota
	screenManage
)

var targetAbbrev = map[string]string{
	"agents": "ag",
	"codex":  "co",
	"claude": "cl",
	"hermes": "he",
}

type model struct {
	repoDir string
	skills  []Skill
	targets []Target

	// statuses[skillIdx][targetIdx]
	statuses [][]Status

	screen  screen
	cursor  int // skill index, list screen
	manage  int // skill index being managed
	tcursor int // target index, manage screen

	confirming bool
	message    string
	messageErr bool

	width, height int
	ready         bool
}

func newModel(repoDir string, skills []Skill, targets []Target) model {
	m := model{repoDir: repoDir, skills: skills, targets: targets}
	m.refreshAll()
	return m
}

func (m *model) refreshAll() {
	m.statuses = make([][]Status, len(m.skills))
	for i, s := range m.skills {
		row := make([]Status, len(m.targets))
		for j, t := range m.targets {
			row[j], _ = t.status(s)
		}
		m.statuses[i] = row
	}
}

func (m *model) refreshSkill(i int) {
	row := make([]Status, len(m.targets))
	for j, t := range m.targets {
		row[j], _ = t.status(m.skills[i])
	}
	m.statuses[i] = row
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.ready = true
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.screen {
	case screenList:
		return m.handleListKey(msg)
	case screenManage:
		return m.handleManageKey(msg)
	}
	return m, nil
}

func (m model) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "j", "down":
		if len(m.skills) > 0 {
			m.cursor = min(m.cursor+1, len(m.skills)-1)
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "g", "home":
		m.cursor = 0
	case "G", "end":
		if len(m.skills) > 0 {
			m.cursor = len(m.skills) - 1
		}
	case "r":
		m.refreshAll()
		m.message, m.messageErr = "refreshed install status", false
	case "enter":
		if len(m.skills) > 0 {
			m.screen = screenManage
			m.manage = m.cursor
			m.tcursor = 0
			m.confirming = false
			m.message = ""
		}
	}
	return m, nil
}

func (m model) handleManageKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.confirming {
		switch msg.String() {
		case "y":
			t := m.targets[m.tcursor]
			s := m.skills[m.manage]
			if err := t.uninstall(s); err != nil {
				m.message, m.messageErr = err.Error(), true
			} else {
				m.message, m.messageErr = fmt.Sprintf("%s: uninstalled", t.Name), false
			}
			m.refreshSkill(m.manage)
			m.confirming = false
		case "n", "esc":
			m.confirming = false
			m.message = ""
		case "ctrl+c":
			return m, tea.Quit
		}
		return m, nil
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "q":
		m.screen = screenList
		m.message = ""
	case "j", "down":
		m.tcursor = min(m.tcursor+1, len(m.targets)-1)
	case "k", "up":
		if m.tcursor > 0 {
			m.tcursor--
		}
	case " ", "enter":
		t := m.targets[m.tcursor]
		s := m.skills[m.manage]
		switch m.statuses[m.manage][m.tcursor] {
		case StatusNotInstalled:
			if err := t.install(s); err != nil {
				m.message, m.messageErr = err.Error(), true
			} else {
				m.message, m.messageErr = fmt.Sprintf("%s: installed", t.Name), false
			}
			m.refreshSkill(m.manage)
		case StatusInstalled:
			m.confirming = true
			m.message = ""
		case StatusConflict:
			m.message, m.messageErr = fmt.Sprintf("%s: something else is already there — resolve it, or run install-skill --force from a shell", t.Name), true
		case StatusAgentMissing:
			m.message, m.messageErr = fmt.Sprintf("%s: agent not detected on this machine", t.Name), true
		}
	}
	return m, nil
}

func (m model) View() string {
	if !m.ready {
		return ""
	}
	var out string
	switch m.screen {
	case screenManage:
		out = m.viewManage()
	default:
		out = m.viewList()
	}
	return clampWidth(out, m.width)
}

// clampWidth is the last line of defense for the bounded-render invariant:
// whatever screen-specific layout math produced, no emitted line may exceed
// the terminal's actual width. ANSI-aware, so it's safe on already-styled
// lines (see modern-tui's text-and-width reference).
func clampWidth(s string, w int) string {
	if w <= 0 {
		return s
	}
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if lipgloss.Width(line) > w {
			lines[i] = lipgloss.NewStyle().MaxWidth(w).Render(line)
		}
	}
	return strings.Join(lines, "\n")
}

func (m model) viewList() string {
	var b strings.Builder

	b.WriteString(styleAccent.Render(iconMark) + " " + gradientText("skillz", titleGradient, true) + "\n")
	b.WriteString(gradientRule(m.width, titleGradient) + "\n")

	installed, total := m.coverage()
	frac := 0.0
	if total > 0 {
		frac = float64(installed) / float64(total)
	}
	skillCount := fmt.Sprintf("%d skill%s", len(m.skills), plural(len(m.skills)))
	counts := fmt.Sprintf(" %d/%d installs", installed, total)
	// The meter is decoration on top of skillCount/counts, which always
	// render in full - so it's the one part of this line allowed to shrink
	// (down to skipping it) rather than get clamped mid-glyph.
	meterW := m.width - lipgloss.Width(skillCount) - 3 - lipgloss.Width(counts)
	if meterW > 18 {
		meterW = 18
	}
	line := styleDim.Render(skillCount)
	if meterW >= 4 {
		line += "   " + meter(frac, meterW) + styleDim.Render(counts)
	}
	b.WriteString(line + "\n\n")

	nameW := 20
	if m.width < 70 {
		nameW = 12 // still fits the longest current skill name at a glance
	}
	descW := m.width - nameW - 24
	if descW < 6 {
		descW = 6
	}

	for i, s := range m.skills {
		name := padCells(s.Name, nameW)
		desc := padCells(truncateCells(s.Description, descW), descW)
		statusCol := m.statusColumn(i)

		line := name + " " + desc + " " + styleDivider.Render("│") + " " + statusCol
		if i == m.cursor {
			b.WriteString(styleRail.Render(iconFocus+" ") + styleRowSelected.Render(line) + "\n")
		} else {
			b.WriteString("  " + styleRow.Render(line) + "\n")
		}
	}

	if len(m.skills) <= 3 {
		note := truncateCells("More skills are on the way — this repo grows over time.", m.width-2)
		b.WriteString("\n  " + styleFaint.Render(note) + "\n")
	}

	legend := truncateCells("ag agents   co codex   cl claude   he hermes    ● installed   ○ not installed   ! conflict   - agent missing", m.width)
	b.WriteString("\n" + styleFooter.Render(legend) + "\n")

	if m.message != "" {
		b.WriteString(renderMessage(m.message, m.messageErr) + "\n")
	}

	b.WriteString("\n" + keyHints([2]string{"↵", "manage"}, [2]string{"r", "refresh"}, [2]string{"q", "quit"}))
	return b.String()
}

// coverage counts installed slots against every slot whose agent is
// actually present on this machine (an agent that isn't installed at all
// shouldn't count against the bar).
func (m model) coverage() (installed, total int) {
	for _, row := range m.statuses {
		for _, st := range row {
			if st == StatusAgentMissing {
				continue
			}
			total++
			if st == StatusInstalled {
				installed++
			}
		}
	}
	return installed, total
}

func (m model) statusColumn(skillIdx int) string {
	parts := make([]string, len(m.targets))
	for j, t := range m.targets {
		parts[j] = styleDim.Render(targetAbbrev[t.Name]) + statusGlyph(m.statuses[skillIdx][j])
	}
	return strings.Join(parts, "  ")
}

func (m model) viewManage() string {
	s := m.skills[m.manage]

	panelWidth := m.width - 6
	if panelWidth > 78 {
		panelWidth = 78
	}
	// 46 is the smallest panel that still fits a target row's fixed budget
	// (2 prefix + 8 label + 1 + pathMin + 2 + 18 status) without forcing an
	// internal word-wrap; below this a genuinely tiny terminal just gets
	// hard-clamped by clampWidth instead of laid out prettily.
	if panelWidth < 46 {
		panelWidth = 46
	}
	contentWidth := panelWidth - 4 // stylePanel's own Padding(1, 2)

	installed, total := 0, 0
	for _, st := range m.statuses[m.manage] {
		if st == StatusAgentMissing {
			continue
		}
		total++
		if st == StatusInstalled {
			installed++
		}
	}
	frac := 0.0
	if total > 0 {
		frac = float64(installed) / float64(total)
	}

	var body strings.Builder
	body.WriteString(styleAccent.Render(iconMark) + " " + styleName.Render(s.Name) + "\n")
	body.WriteString(styleFaint.Render(truncateCells(s.Description, contentWidth)) + "\n")
	body.WriteString(meter(frac, 16) + styleDim.Render(fmt.Sprintf(" %d/%d installed", installed, total)) + "\n")
	body.WriteString(styleDivider.Render(strings.Repeat("─", contentWidth)) + "\n\n")

	for j, t := range m.targets {
		status := m.statuses[m.manage][j]
		dest, _ := t.dest(s)

		label := padCells(t.Name, 8)
		// Budget every fixed piece explicitly so the row can never exceed
		// contentWidth regardless of which status label lands here - the
		// longest is Status.String()'s "agent not found" (16 cells), and
		// every row (focused or not) carries a 2-cell leading glyph/indent.
		const prefixW, labelW, statusColW = 2, 8, 1 + 1 + len("agent not found")
		pathW := contentWidth - prefixW - labelW - 1 - 2 - statusColW
		if pathW < 6 {
			pathW = 6
		}
		path := styleDim.Render(padCells(truncateCells(dest, pathW), pathW))
		glyph := statusGlyph(status) + " " + statusLabelStyled(status)

		var row string
		if j == m.tcursor {
			row = styleAccent.Bold(true).Render(iconFocus+" ") + styleName.Render(label) + " " + path + "  " + glyph
		} else {
			row = "  " + label + " " + path + "  " + glyph
		}
		body.WriteString(row + "\n")
	}

	body.WriteString("\n")
	if m.confirming {
		t := m.targets[m.tcursor]
		warnBar := lipgloss.NewStyle().Foreground(colorWarn).Bold(true).Reverse(true).
			Render(fmt.Sprintf(" Uninstall %s from %s? ", s.Name, t.Name))
		body.WriteString(warnBar + "  " + keyHints([2]string{"y", "confirm"}, [2]string{"n", "cancel"}))
	} else if m.message != "" {
		body.WriteString(renderMessage(m.message, m.messageErr))
	} else {
		body.WriteString(keyHints([2]string{"space", "toggle"}, [2]string{"esc", "back"}))
	}

	panel := stylePanel.Width(panelWidth).Render(body.String())
	return "\n" + lipgloss.PlaceHorizontal(m.width, lipgloss.Center, panel)
}

func statusLabelStyled(s Status) string {
	switch s {
	case StatusInstalled:
		return styleGood.Render(s.String())
	case StatusConflict, StatusAgentMissing:
		return styleWarn.Render(s.String())
	default:
		return styleDim.Render(s.String())
	}
}

func renderMessage(msg string, isErr bool) string {
	if isErr {
		return styleErr.Render("✘ " + msg)
	}
	return styleGood.Render("✔ " + msg)
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
