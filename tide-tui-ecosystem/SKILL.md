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

A family of terminal apps under `github.com/allisonhere`, separate repos rather than a
monorepo — so nothing enforces consistency automatically. That's this skill's job.

| Repo | What it is | Stack |
|------|------------|-------|
| `tide` | RSS reader with AI summaries | Go + Bubble Tea |
| `tidemail` | keyboard-first email client | Go + Bubble Tea |
| `tideui` | shared multi-pane TUI toolkit — **import it, don't copy it** | Go + Bubble Tea + Lipgloss |
| `ripple` | soft-wrapping multi-line text editor component (has a vim mode) — a component, not an app | Go + Bubble Tea |
| `TideFTP` | FTP/SFTP client (private) | Go + Bubble Tea |
| `whatthedock` | Docker manager TUI | Go + Bubble Tea + tideui |
| `zellit` | Zellij theme editor — **Rust + Ratatui, not Bubble Tea** | Rust |

Adjacent, same house style, worth checking for prior art before writing a new primitive:
`anote` (quick notes TUI), `alogi` (log viewer with AI analysis),
`ratatui-color-picker` (Rust), `cli-spinners-GO-library`.

`whatthedock` is the fullest example of an app built *on* `tideui` — read its
`internal/ui/` before reinventing panes, meters, or dashboards.

## Local repo map

Expected local clone names are not perfectly normalized:

| Repo | Local directory name to check first | Import/module cue |
|------|-------------------------------------|-------------------|
| `tide` | `tide` | app, owns its UI state |
| `tidemail` | `tidemail` | app, duplicated primitives may exist |
| `tideui` | `tideui` | shared UI module; import before copying |
| `ripple` | `ripple` | text editor component |
| `TideFTP` | `tideftp` | private app, lowercase local dir |
| `whatthedock` | `whatthedock` | app using `tideui` heavily |
| `zellit` | `zellit` | Rust/Ratatui, not Go/Bubble Tea |

Default to `tideui` for panes, rows, overlays, theme roles, scrolling, ratios, and
compact meters. Default to `ripple` for editable multi-line text. Copying a primitive is
only reasonable when the target repo already owns a deliberately divergent version.

## Shared code, not shared modules

Some files are duplicated rather than imported — confirmed: `ansi_wave.go` exists in
both `tide` and `tidemail`. Fixing one is half a fix.

Check the siblings before calling such a change done. Local clones first:

```bash
ls -d ~/Projects/tide ~/Projects/tidemail ~/Projects/tideui ~/Projects/tideftp ~/Projects/whatthedock 2>/dev/null
grep -rl 'distinctiveFunctionName' ~/Projects/tide* ~/Projects/whatthedock 2>/dev/null
```

Not every repo is cloned on every machine (`tide` and `ripple` frequently aren't). Then:

```bash
gh api repos/allisonhere/tide/contents --jq '.[].name' | grep -i ansi_wave
gh search code --owner allisonhere --filename ansi_wave.go
```

**Caveat:** `gh search code` is index-backed and incomplete — it returned only `tide`
for `ansi_wave.go` while `tidemail` demonstrably has the file. Treat a search miss as
"unknown", never as "not duplicated"; confirm per repo with the contents API.

`tideui` is the one repo meant to be a real module dependency. If you are writing pane
chrome, layout math, scrolling, theming, or modal overlays, check `tideui` first — it
already exposes a `Renderer` (panes/rows/blocks/overlays), `PaneRatio` (bounded split
ratios), `PaneScroller` (offset + clamping), a `Theme` set with contrast enforcement, and
a theme picker with live preview. `ripple` is the same deal for text editing.

## Docs convention

These repos favor a `STATUS.md` (current state, working/not-working feature list) and
sometimes a `PLAN.md` alongside `README.md` and `CLAUDE.md`. When a change materially
affects what's working, update `STATUS.md` — it's the quick-scan doc, don't let it
drift out of sync with reality.

## Release pattern

Verified per repo — there is no single family rule:

- `tide`: `release.sh` + `install.sh` at root
- `murmur`-style: `install.sh` only
- `ripple`: neither — it is imported, not installed
- `tideui`: no release scripts; consumers pin it via `go.mod`

Check `ls release.sh install.sh Makefile .goreleaser.y*ml` before assuming a flow. The
general Go release/versioning checklist lives in the `go-cli-workflow` skill — including
the `git describe --match 'v*'` tag-series trap. Don't duplicate it here.

## When starting new Tide-family work

If the user is adding a new Bubble Tea TUI app that feels like it belongs to this
family (terminal-first, keyboard-driven, in the allisonhere namespace), default to:
- Building UI chrome on `tideui` rather than reinventing panes/scrolling/theming.
- Following the `STATUS.md` + `CLAUDE.md` doc pattern from day one.
- Matching the general Go CLI conventions in `go-cli-workflow`.

## See also

- `modern-tui` — portable TUI design/interaction/architecture craft (layouts, focus,
  palettes, width safety, bounded rendering). Use it for *how to build the UI*.
- `go-cli-workflow` — preflight commands, versioning, multi-module traps, releases.
  Use it for *how to build and ship the binary*.

This skill holds only the cross-repo facts neither of those can know: who imports whom,
what is duplicated, and which docs each repo keeps.
