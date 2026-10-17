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

Also duplicated or ported between repos (verified 2026-10-15):

- **Drop-shadow compositing**: `tidemail/internal/ui/util.go` (`blendShadowRect`, `overlayAt`) and
  `color.go` (`shadowColor`, `blendBg`) were ported into `tide/internal/ui/shadow.go` (1×1 offset,
  `display.shadow`). `tideui/layout.go` has its own copy with a 2×1 offset (`ModalShadow`).
- **Settings → Changelog**: `internal/ui/settings_changelog.go` in both `tidemail` and `tide`,
  fed by an embedded `CHANGELOG.md` (`//go:embed` in `main.go`).
- **`internal/omarchy/omarchy.go`** (theme detection) is the same file in both. The Omarchy *bar*
  code (`tide/internal/omarchy/bar.go`, `tide --bar`) exists only in `tide`.
- **Headless CLI**: `tidemail` has `internal/cli` (`messages list --json`, `skill`,
  `install-skill`); `tide` has only `--refresh`, `--unread-count`, `--bar`, `--help` (`cli.go`).
- **Windows**: `tide` and `tidemail` both ship experimental Windows zips; per the owner,
  `tideftp` and `whatthedock` still have Windows problems.

`tideui` is the one repo meant to be a real module dependency. If you are writing pane
chrome, layout math, scrolling, theming, or modal overlays, check `tideui` first — it
already exposes a `Renderer` (panes/rows/blocks/overlays), `PaneRatio` (bounded split
ratios), `PaneScroller` (offset + clamping), a `Theme` set with contrast enforcement, and
a theme picker with live preview. `ripple` is the same deal for text editing.

**Check `tideui`'s tags before assuming an API is available.** `tidemail` pins
`tideui v0.2.2` and uses it for `Theme`, the palettes, and `PaneScroller` (in `cliform`). The colour/contrast helpers
(`ContrastRatio`, `ReadableText`, `MutedText`, `AdjustLightness`, `MixColors`,
`AccentReadableOn`, `SurfaceForRatio`, `Hue`, `ShiftHue`) are exported on `tideui`'s
`main` but were in no tagged release when last checked, so `tidemail/docs/tideui-migration.md`
(which still calls them unexported) is stale, and `tideftp` still carries its own copy.
Verify with `git tag --contains <commit>` in `~/Projects/tideui`. `tideui` has no
changelog viewer, scene canvas, or double-click helper; those live only in `tidemail`.

**Other shared libraries.** `ripple` is the text editor (`tidemail`, `tidesms`, `tidegit`,
`tidedeck`, `whatthedock` import it; `tidemail` was still pinned at v0.2.0 when checked). Its
v0.3.0 adds `StyleKey`/`Style` hooks for host-driven inline highlighting, so decorations such
as spelling underlines need no editor change. `tidespell` (v0.1.0, created 2026-10-15) is an
offline English spell checker: embedded SCOWL word lists, no cgo, `Check` returns misspelled
spans as rune offsets that feed ripple's `StyleKey`, `Suggest` ranks corrections. No app uses
it yet. Put reusable, UI-free logic in a small module like these instead of copying it between
apps.

## Docs convention

These repos favor a `STATUS.md` (current state, working/not-working feature list) and
sometimes a `PLAN.md` alongside `README.md` and `CLAUDE.md`. When a change materially
affects what's working, update `STATUS.md` — it's the quick-scan doc, don't let it
drift out of sync with reality.

## User-facing docs and wiki

When behavior changes, update the docs at the same time as the code. Check the repo contributor instructions, then keep its primary guide, README, current-status page, in-app help, and shortcut reference consistent wherever they describe that behavior. Search for the old wording so secondary help screens do not keep teaching the previous flow.

A GitHub wiki is a separate repository and does not update when the source repo changes. Inspect its page list and edit the relevant pages and navigation index alongside the source docs. Link wiki explanations to the source revision or release they describe, and verify the published page after pushing. If the wiki is unavailable, finish the source docs and record the wiki update as outstanding rather than implying it was published.

## Release pattern

Verified per repo — there is no single family rule:

- `tide`: `release.sh` + `install.sh` at root, and `deploy.sh` + `deploy.conf`
- `tidemail`, `tidedeck`, `tidesms`: `deploy.sh` + `deploy.conf` (see below)
- `murmur`-style: `install.sh` only
- `ripple`: neither — it is imported, not installed
- `tideui`, `tideftp`: no `deploy.sh`; `tideftp` consumes `tideui` via `go.mod`, and
  `tideui` is released by tagging

Check `ls deploy.sh deploy.conf release.sh install.sh Makefile .goreleaser.y*ml` before
assuming a flow.

### `deploy.sh` — the shared Tide release console

`deploy.sh` is the same ~80 KB script in every repo that has one (its `deploy.conf`
header says it is identical across Tide repos); only `deploy.conf` differs (app name,
binary, repo, AUR package, `RELEASE_NOTES=changelog`). Facts verified in `tidemail`:

- **It releases whatever branch is checked out** (`git rev-parse --abbrev-ref HEAD`).
  Tagging from a feature branch leaves `main` behind the release (v1.4.0 was cut from
  `feat/context-menu` while `main` was still at v1.3.1). Switch to `main` and merge
  first, or fast-forward `main` to the tag afterwards (`git merge --ff-only vX.Y.Z`).
- **The commit step runs `git add -A`** after a confirmation that lists the changes, and
  commits them as `chore: release $VERSION`. Any untracked or half-finished file in the
  tree rides along — an unrelated assessment doc was committed this way. Run
  `git status` first; commit or move strays out; read the confirmation list.
- **Release notes come from `CHANGELOG.md`.** The "Release notes" step renames
  `## Unreleased` to `## <version>` and opens a fresh empty `## Unreleased` above it.
  If `## <version>` already exists it skips promotion. The pre-flight only *warns* when
  Unreleased is empty. Put new entries under `## Unreleased`. That step only exists when
  `deploy.conf` sets `RELEASE_NOTES="changelog"` (`tidemail`); `tide` uses `"generated"` (GitHub's
  automatic notes), so its pipeline has no notes step and the heading is renamed by hand. A dry
  run still rebuilds the tracked `dist/` binaries.
- **CI builds from the tag, not `main`.** `.github/workflows/release.yml` builds from
  `github.ref_name` and publishes the changelog section matching the tag, plus
  `SHA256SUMS` for the in-app updater — so never move or re-push a published tag; ship a
  patch version instead.
- `ci.yml` runs gofmt, build, vet, `go test -race`, and `golangci-lint` at `version:
  latest` — a newer linter than yours can fail a release that passes locally. The
general Go release/versioning checklist lives in the `go-cli-workflow` skill — including
the `git describe --match 'v*'` tag-series trap. Don't duplicate it here.

## Async state in Bubble Tea

Commands run concurrently and often carry snapshots of model state. When a command saves or mutates state, give the operation a session identity and revision, serialize writes that share a durable record, and let only the acknowledgment for the current revision clear dirty state. Closing, discarding, reopening, and undoing work must also define what happens to outstanding commands. Check both the command result and the visible/durable state in a regression test.

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
