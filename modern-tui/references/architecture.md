# TUI Architecture

## General Separation

Separate at least these concerns:

- domain/application state
- UI/focus state
- rendering/view construction
- side effects and external services
- persistence/configuration

Rendering should be close to a pure projection of state.

## Bubble Tea / Lipgloss

Use the Elm-style contract deliberately:

- `Model` owns application and UI state
- `Update` handles messages and returns commands
- `View` renders current state
- `tea.Cmd` performs asynchronous work

Do not perform slow network, filesystem, subprocess, or API calls directly in `View` or synchronously inside event handling.

Represent long-running work with explicit messages such as started/succeeded/failed and render those states.

Keep reusable view primitives stateless when possible. A shared renderer/toolkit should accept content, dimensions, focus flags, and theme, then return bounded output. Let the application own persistence, commands, and service state.

## Ratatui / Crossterm

Keep event collection, app state updates, rendering, and async/service work distinct.

A typical architecture should have:

- terminal/event loop
- application state struct
- input/action mapping
- update/reducer-style state transitions
- renderer/components
- channels/tasks for slow work

Avoid embedding service calls inside widget rendering.

## Action Layer

For nontrivial applications, map raw key events to semantic actions. This makes remapping, command palettes, tests, and contextual command availability easier.

Example semantic actions:

- `MoveDown`
- `FocusNextPane`
- `OpenSelected`
- `RefreshAll`
- `ToggleUnread`
- `OpenCommandPalette`

## Background Work

Any operation that may noticeably block should leave the UI loop:

- HTTP requests
- SSH
- Docker/Podman APIs
- large directory scans
- database migrations/queries
- AI inference
- package/update checks
- subprocess execution

Feed completion back through messages/events and redraw only when state changes or animation requires it.

## Resize Handling

Treat terminal dimensions as live state. Clamp all calculations. Never subtract fixed chrome sizes without guarding underflow/negative dimensions.

Rendering should have defined behavior for tiny dimensions.

## Reusable Components

Good candidates:

- pane shell
- status bar
- modal/soft panel
- theme picker
- command palette
- list row/block
- ratio helper
- scrolling helper
- notification/toast
- confirm dialog
- compact meters/sparklines

Do not extract a component merely because two functions look similar. Extract when behavior and design semantics are shared.

## Persistence

Persist user-facing choices that reasonably form part of the workspace experience: theme, density, pane ratios, optional display toggles, and perhaps last focus/view if consistent with product intent.

Keep secrets separate from casually displayed configuration and avoid accidental logging.

## Testing

Test state transitions independently from rendering where possible.

For rendering, verify:

- output is bounded
- tiny dimensions do not panic
- selected/focused states are distinguishable
- modal positioning stays valid
- truncation does not corrupt Unicode
- themes do not produce invisible foreground/background combinations
