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

### Minimal Bubble Tea spine

```go
type Model struct {
    w, h    int          // live terminal size
    ready   bool         // first WindowSizeMsg seen
    loading bool
    err     error
    items   []Item
    sel     int
    offset  int
}

type itemsLoadedMsg struct {
    items []Item
    err   error
}

func loadItems(client *Client) tea.Cmd {
    return func() tea.Msg { // runs off the UI loop
        items, err := client.List(context.Background())
        return itemsLoadedMsg{items: items, err: err}
    }
}

func (m Model) Init() tea.Cmd { return loadItems(m.client) }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.w, m.h, m.ready = msg.Width, msg.Height, true
        m.clamp() // re-clamp selection + offset for the new body height
        return m, nil
    case itemsLoadedMsg:
        m.loading = false
        m.items, m.err = msg.items, msg.err
        m.clamp()
        return m, nil
    case tea.KeyMsg:
        return m.handleKey(msg)
    }
    return m, nil
}

func (m Model) View() string {
    if !m.ready {
        return "" // never render before a size is known
    }
    // pure projection of state; no I/O, no panics, bounded to m.w x m.h
    return m.render()
}
```

`Init` starts work; `Update` is the only mutator; `View` is pure. A `tea.Cmd` is the *only* place slow work belongs.

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

### Minimal Ratatui spine

```rust
enum Event { Key(KeyEvent), Resize(u16, u16), Tick, Items(Result<Vec<Item>, String>) }

fn run(terminal: &mut Terminal<impl Backend<Error = io::Error>>, rx: &Receiver<Event>, app: &mut App) -> io::Result<()> {
    loop {
        terminal.draw(|f| render(f, app))?;          // pure: state -> frame
        match rx.recv() {                            // one place events enter
            Ok(Event::Key(k))    => if app.on_key(k).is_quit() { return Ok(()); },
            Ok(Event::Resize(..))=> app.clamp(),     // size is live state
            Ok(Event::Tick)      => app.on_tick(),
            Ok(Event::Items(r))  => app.on_items(r), // completion of async work
            Err(_)               => return Ok(()),   // producers gone
        }
    }
}
```

Input polling, async producers, and ticks all feed one channel; `App` owns state; `render` reads it. Never call a service from inside `render`.

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
