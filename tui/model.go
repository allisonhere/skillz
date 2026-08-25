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

	subtitle := fmt.Sprintf("%d skill%s available", len(m.skills), plural(len(m.skills)))
	b.WriteString(styleTitle.Render("skillz") + "  " + styleDim.Render(subtitle) + "\n\n")

	nameW := 20
	descW := m.width - nameW - 22
	if descW < 10 {
		descW = 10
	}

	for i, s := range m.skills {
		name := padCells(s.Name, nameW)
		desc := padCells(truncateCells(s.Description, descW), descW)
		statusCol := m.statusColumn(i)

		line := name + " " + desc + " " + statusCol
		if i == m.cursor {
			b.WriteString(styleRail.Render("▌") + styleRowSelected.Render(line) + "\n")
		} else {
			b.WriteString(" " + styleRow.Render(line) + "\n")
		}
	}

	if len(m.skills) <= 3 {
		b.WriteString("\n  " + styleFaint.Render("More skills are on the way — this repo grows over time.") + "\n")
	}

	b.WriteString("\n" + styleFooter.Render("ag agents  co codex  cl claude  he hermes    ● installed  ○ not installed  ! conflict  - agent missing") + "\n")

	if m.message != "" {
		b.WriteString(renderMessage(m.message, m.messageErr) + "\n")
	}

	b.WriteString("\n" + styleFooter.Render("↵ manage    r refresh    q quit"))
	return b.String()
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

	var body strings.Builder
	body.WriteString(styleName.Render(s.Name) + "\n")
	body.WriteString(styleDim.Render(truncateCells(s.Description, 64)) + "\n\n")

	for j, t := range m.targets {
		status := m.statuses[m.manage][j]
		dest, _ := t.dest(s)

		label := padCells(t.Name, 8)
		path := styleDim.Render(truncateCells(dest, 44))
		glyph := statusGlyph(status) + " " + statusLabelStyled(status)

		row := label + " " + path + "  " + glyph
		if j == m.tcursor {
			row = styleAccent.Render("▸ ") + row
		} else {
			row = "  " + row
		}
		body.WriteString(row + "\n")
	}

	if m.confirming {
		t := m.targets[m.tcursor]
		body.WriteString("\n" + styleWarn.Render(fmt.Sprintf("Uninstall %s from %s? y/n", s.Name, t.Name)))
	} else if m.message != "" {
		body.WriteString("\n" + renderMessage(m.message, m.messageErr))
	} else {
		body.WriteString("\n" + styleFooter.Render("space toggle   esc back"))
	}

	panelWidth := m.width - 6
	if panelWidth > 76 {
		panelWidth = 76
	}
	if panelWidth < 30 {
		panelWidth = 30
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
