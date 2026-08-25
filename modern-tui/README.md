# modern-tui

An agent skill for designing, building, and reviewing keyboard-first terminal
applications — multi-pane layouts, focus/selection, command palettes, themes,
compact terminal graphs — without forcing a framework change on the project.

Portable `SKILL.md` bundle (Agent Skills pattern), usable by Codex, Claude Code,
Hermes, and anything else that scans a skills directory.

## When it fires

Triggers on requests like *design a TUI*, *build a terminal UI*, *review this
TUI*, *make this TUI modern*, *add a command palette*, *add terminal graphs*, or
any work on a keyboard-first terminal app in Bubble Tea/Lipgloss,
Ratatui/Crossterm, or a comparable framework.

Framework-agnostic by design: it detects and preserves the project's existing
stack, and only recommends a stack for greenfield work.

## Contents

```text
modern-tui/
├── SKILL.md                          # principles + workflow (the entry point)
├── references/
│   ├── design-language.md            # layouts, hierarchy, density, borders, themes, modals
│   ├── interaction.md                # navigation, command palette, help, search, forms, mouse
│   ├── architecture.md               # Bubble Tea / Ratatui structure, async work, resize handling
│   ├── graphs.md                     # sparklines, meters, thresholds, missing data
│   └── review-checklist.md           # systematic design/code review pass
├── templates/
│   └── repo-mining-prompt.md         # prompt to mine existing TUI repos for house patterns
└── README.md
```

`SKILL.md` stays lean and names the reference to read for a given task — load
only what the task needs rather than the whole bundle.

## What it optimizes for

- information-rich over wastefully spacious
- keyboard-first without excluding mouse users
- whitespace and hierarchy over boxing every widget
- focus and selection that never rely on color alone
- rendering as a projection of state, with slow work off the UI loop
- bounded rendering that survives tmux splits, 80×24, and tiny windows

## Install

From this repo (creates symlinks, so edits here are live immediately):

```bash
~/Projects/skillz/install-skill modern-tui            # agents + codex + claude
~/Projects/skillz/install-skill modern-tui --target codex
~/Projects/skillz/install-skill modern-tui --target hermes   # opt-in, copy-based
```

Targets, destinations, and the symlink-vs-copy rules are documented in the
[repo README](../README.md). Hermes is a **copy**, so re-run the `--target
hermes` install after editing the skill.

Manual install without the script: copy or symlink this whole directory (the
folder containing `SKILL.md`) into the agent's skills location — e.g.
`~/.claude/skills/modern-tui` or `~/.hermes/skills/software-development/modern-tui`.

## Verify it loaded

- Codex / Claude Code: start a session and ask it to *review this TUI layout*;
  the skill should be listed among available skills.
- Hermes: `hermes skills list` (or start a new session) and confirm
  `modern-tui` appears under `software-development`.

## Editing

`~/Projects/skillz/modern-tui` is the canonical source. Keep new detail in
`references/` and keep `SKILL.md` short; add a reference file only when the
topic is substantial enough to justify one. Do not put machine-specific paths,
credentials, or single-application quirks in the skill body — application
conventions belong in the Tide-family section of `SKILL.md` if they are
genuinely house style.
