---
name: modern-tui
description: Design, build, or review a keyboard-first TUI (terminal UI) — multi-pane layouts, command palettes, focus/selection, themes, compact terminal graphs — in Bubble Tea/Lipgloss, Ratatui/Crossterm, or a comparable framework.
version: 1.1.0
license: MIT
platforms: [linux, macos, windows]
metadata:
  hermes:
    tags: [tui, terminal, ui, ux, bubble-tea, lipgloss, ratatui, rust, go]
    category: software-development
---

# Modern TUI

Design and implement polished, keyboard-first terminal applications with strong information hierarchy, responsive layouts, predictable interaction, and restrained visual chrome.

## Core Principle

Treat the terminal as a real application surface, not a collection of bordered widgets. Optimize for clarity, speed, density, discoverability, and character.

Prefer:

- information-rich over wastefully spacious
- consistent over clever
- subtle over flashy
- keyboard-first without excluding mouse users
- whitespace and hierarchy over excessive borders
- compact visualizations over oversized charts
- responsive stateful applications over blocking command wrappers

Do not force a framework change. Detect and preserve the project's existing stack unless migration is explicitly requested.

## Start Here

| Task | Read |
|------|------|
| New TUI from scratch | `references/architecture.md`, then `references/design-language.md` |
| Restyle / relayout an existing TUI | `references/design-language.md` |
| Keys, palette, help, forms, scrolling | `references/interaction.md` |
| Alignment, truncation, emoji/CJK, padding bugs | `references/text-and-width.md` |
| Panic wrecks the terminal, Ctrl-Z, clipboard, `$EDITOR`, logging | `references/terminal-lifecycle.md` |
| No color, `TERM=dumb`, piped output, tmux, ASCII fallback | `references/capabilities.md` |
| Slow redraw, huge lists, high-frequency updates | `references/performance.md` |
| Dashboards, metrics, sparklines, meters | `references/graphs.md` |
| Writing tests for a TUI | `references/testing.md` |
| Reviewing someone's TUI | `references/review-checklist.md` |

## Framework Selection

When extending an existing project, use its current framework and conventions.

For a new project:

- Prefer Go + Bubble Tea + Lipgloss when the project benefits from Elm-style message/update/view architecture, rapid UI iteration, or compatibility with the Tide-family ecosystem.
- Prefer Rust + Ratatui + Crossterm when Rust is already desired, low-level control matters, or the surrounding application is Rust.
- Consider an existing reusable UI layer before reimplementing primitives. In Tide-family Go applications, inspect whether `tideui` fits the requirement.

Do not rewrite a working TUI into another language merely because another framework is familiar.

## Workflow

1. Inspect the existing project before proposing UI changes.
2. Identify framework, state model, event loop, layout system, theme system, key routing, async/background work, and terminal-size handling.
3. Preserve useful conventions already established by the application.
4. Sketch the information architecture before writing rendering code.
5. Define focus, selection, navigation, overlays, and destructive-action behavior explicitly.
6. Implement the smallest coherent change.
7. Verify keyboard behavior, narrow-terminal behavior, theme contrast, and redraw responsiveness.
8. Run relevant tests and add targeted UI/state tests where practical.

## Visual Language

Use a contemporary terminal aesthetic:

- clear pane hierarchy
- compact rows
- meaningful whitespace
- restrained borders
- strong but tasteful focus indication
- quiet metadata
- semantic color roles
- concise status/footer hints
- soft overlays or modals
- Unicode where appropriate with graceful fallback

Avoid:

- boxing every widget automatically
- gratuitous ASCII decoration
- giant headings that consume vertical space
- excessive permanent shortcut legends
- using color as the only status signal
- decorative animation that makes interaction slower
- layouts that break when placed in tmux/zellij splits

For detailed guidance, read `references/design-language.md`.

## Layouts

Choose layout based on information relationships, not habit.

Strong defaults include:

- sidebar + list + detail
- three columns for peer-to-peer inspection
- sidebar + vertically stacked work panes
- tabbed mode for narrow terminals
- floating/overlay panels for transient tasks

Allow adjustable ratios when panes have competing information needs. Persist user-adjusted ratios when the application already persists display preferences.

Collapse nonessential metadata before making the primary task unusable.

## Focus and Selection

Make focus unambiguous. Distinguish:

- focused pane
- selected row
- active item
- disabled item
- hovered item, if mouse support exists

Never require color alone to communicate these states.

When borders are part of the product's visual language, a full focused-pane border is valid. When borders add noise without communicating focus or grouping, prefer whitespace, headings, rails, or background contrast.

## Keyboard Interaction

Support arrows even when Vim keys are available.

Common defaults:

- `j` / `k` and `↓` / `↑`: move within a list
- `h` / `l` and `←` / `→`: move between adjacent panes when spatially appropriate
- `Tab` / `Shift-Tab`: cycle focus
- `Enter`: open, activate, or confirm
- `Esc`: cancel, close, or move one level back
- `/`: search or filter
- `?`: contextual help
- `q`: quit from the main surface when appropriate

Do not overload the same key with unrelated actions in the same context.

For applications with many commands, use a searchable command palette. `Ctrl-K` is a strong default when it does not conflict with existing application behavior.

Read `references/interaction.md` before redesigning key routing, command palettes, status hints, forms, or overlays.

## Status Bars and Hints

Keep persistent hints intentionally small. Show only high-value, context-relevant actions.

Prefer a line such as:

`m manage   S settings   / search   ? help`

over a footer containing every command in the application.

Put the full shortcut catalog behind contextual help or a command palette.

## Themes and Accessibility

Centralize semantic theme roles. Prefer names such as:

- background
- foreground
- muted
- accent
- selection
- border
- border-focus
- success
- warning
- danger

Ensure selection and focus remain visible on unusually light and dark themes. When practical, enforce measurable contrast rather than assuming a fixed color delta is sufficient.

Offer ASCII or reduced-Unicode fallbacks when terminal compatibility is a goal.

## Graphs and Metrics

Favor small terminal-native visualizations when they answer the question without opening a dedicated chart view.

Useful one-line forms include:

- sparklines: `▁▂▃▅▄▆▇█`
- filled meters: `██████░░░░`
- segmented meters: `▰▰▰▰▰▱▱▱▱▱`
- occupancy bars: `■■■■■□□□□□`

Read `references/graphs.md` when designing dashboards, resource monitors, Docker views, traffic displays, or other metric-heavy surfaces.

## Architecture

Keep rendering separate from application/business state.

Rendering should consume state; it should not own network clients, persistence, or service logic.

Never block the UI loop with network requests, filesystem scans, AI inference, subprocess execution, Docker calls, or expensive database work. Run slow work asynchronously and feed results back as messages/events.

Handle resize events explicitly. Rendering must remain bounded to the requested terminal dimensions and must not panic on tiny windows.

Read `references/architecture.md` before implementing or restructuring a substantial TUI.

## Destructive Actions

Require clear confirmation for actions such as delete, prune, purge, destructive overwrite, or remote mutation with difficult rollback.

Do not add confirmation to harmless, instantly reversible navigation actions.

## Empty, Loading, and Error States

Never leave unexplained blank panes.

Examples:

- `No containers found`
- `No feeds yet — press a to add one`
- `Refreshing…`
- `Could not connect to Docker daemon`

Keep detailed diagnostics available without making raw stack traces the normal interface.

## Testing and Verification

Before considering a TUI change complete, verify:

- normal keyboard navigation
- arrow-key equivalents where applicable
- focus visibility
- selected-row visibility
- modal open/cancel/confirm behavior
- small terminal behavior
- resize behavior
- long text truncation/wrapping
- empty/loading/error states
- theme contrast
- async responsiveness
- command palette/search behavior if present

Use state-transition tests for interaction logic and snapshot/buffer/golden tests for important rendering where the framework makes this practical.

Read `references/testing.md` for tools, commands, and the bounded-render invariant.

Read `references/review-checklist.md` for a complete design/code review pass.

## Tide-Family Conventions

When working on Tide, TideMail, tideui, or a related application, preserve the established design language unless the task explicitly changes it:

- multi-pane information architecture
- keyboard-first operation
- compact/comfortable density
- theme-aware overlays
- live theme preview where appropriate
- bounded output
- strong visible focus
- application-owned model/state with reusable view primitives
- concise status hints and contextual help

Prefer extending reusable primitives in `tideui` when a behavior is genuinely reusable across applications. Keep application-specific state, commands, persistence, and service integration in the application.

## Additional Resources

Load only what the task needs:

- `references/design-language.md` — layouts, visual hierarchy, themes, borders, density, modals
- `references/interaction.md` — navigation, command palette, help, forms, scrolling, mouse behavior
- `references/architecture.md` — Bubble Tea and Ratatui architecture, async work, state separation, resize handling
- `references/text-and-width.md` — display width, grapheme clusters, ANSI-safe truncation and padding
- `references/terminal-lifecycle.md` — startup/teardown, panics, signals, suspend/resume, clipboard, `$EDITOR`, logging
- `references/capabilities.md` — color profiles, `NO_COLOR`, non-TTY output, Unicode/ASCII degradation, tmux
- `references/performance.md` — frame budget, redraw discipline, list virtualization, update coalescing
- `references/graphs.md` — compact TUI charts and metric patterns
- `references/testing.md` — state-transition tests, golden/snapshot rendering, randomized bounded-render invariants
- `references/review-checklist.md` — systematic review checklist

Use `templates/repo-mining-prompt.md` to have Codex or another repository-aware coding agent extract additional house patterns from existing TUI repositories.
