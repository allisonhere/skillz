---
name: tide-tui-ecosystem
description: Use when editing any Tide-family terminal app (tide, tidemail, tideui, ripple, TideFTP, whatthedock, zellit) — cross-repo facts these separate repos can't show you, from duplicated files to what tideui already provides to per-repo release shape.
version: 1.0.0
license: MIT
platforms: [linux, macos]
metadata:
  hermes:
    tags: [tui, bubble-tea, lipgloss, go, tideui, cross-repo]
    category: software-development
---

# Tide TUI ecosystem

`tide`, `tidemail`, `tideui`, and `ripple` are a related family of Go terminal apps,
all built on Bubble Tea + Lipgloss, under github.com/allisonhere. They're separate
repos, not a monorepo, so nothing enforces consistency automatically — that's this
skill's job.

## Shared code, not shared modules

Some code is duplicated across repos rather than imported as a dependency — notably
`ansi_wave.go`, which exists near-identically in both `tide` and `tidemail`. If you're
fixing a bug or changing behavior in a file like this, check whether the sibling repo
has the same file before assuming the fix is complete. Grep the other Tide repos
locally (`~/Projects/tide`, `~/Projects/tidemail`, etc.) for the same filename or a
distinctive function name before calling the change done.

`tideui` ("Themeable multi-pane terminal UI toolkit for Bubble Tea and Lipgloss
applications") is the one repo of the four meant to be imported as an actual Go
module dependency rather than copy-pasted. If you're duplicating UI chrome, layout,
or theming logic that already exists in `tideui`, prefer depending on it instead of
reimplementing — that's the point of the toolkit existing as its own repo.

## Docs convention

These repos favor a `STATUS.md` (current state, working/not-working feature list) and
sometimes a `PLAN.md` alongside `README.md` and `CLAUDE.md`. When a change materially
affects what's working, update `STATUS.md` — it's the quick-scan doc, don't let it
drift out of sync with reality.

## Release pattern

Distribution is via `release.sh` and/or `install.sh` at the repo root — same
tag-driven versioning as the rest of the user's Go CLI tools (see the
`go-cli-workflow` skill for the general Go release checklist; it applies here too).
Check for a repo-specific `release.sh` before assuming the generic goreleaser/Makefile
flow.

## When starting new Tide-family work

If the user is adding a new Bubble Tea TUI app that feels like it belongs to this
family (terminal-first, keyboard-driven, in the allisonhere namespace), default to:
- Building UI chrome on `tideui` rather than reinventing panes/scrolling/theming.
- Following the `STATUS.md` + `CLAUDE.md` doc pattern from day one.
- Matching the general Go CLI conventions in `go-cli-workflow`.
