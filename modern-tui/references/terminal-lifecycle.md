# Terminal Lifecycle

Everything a TUI turns on, it must turn off — on clean exit, on error, on panic, and on signal.

## Enter / Leave Contract

Track what you enabled and unwind in reverse: alt screen, raw mode, mouse reporting, bracketed paste, focus reporting, cursor visibility, keyboard protocol, and any OSC colors you set.

If you set the terminal background with OSC 11, emit the reset (OSC 111) on the way out. Keep sequence *construction* in a library and *writing* in the application, so tests can assert on strings without touching a real terminal.

## Panics Must Restore the Terminal

An unrestored panic leaves the user in raw mode with no cursor — their shell looks broken.

Rust (Ratatui/Crossterm) — install a hook before the first `enable_raw_mode`:

```rust
use std::io;
use crossterm::{execute, terminal::{disable_raw_mode, LeaveAlternateScreen}, event::DisableMouseCapture};

pub fn install_panic_hook() {
    let previous = std::panic::take_hook();
    std::panic::set_hook(Box::new(move |info| {
        let _ = disable_raw_mode();
        let _ = execute!(io::stdout(), DisableMouseCapture, LeaveAlternateScreen);
        previous(info);
    }));
}
```

Go (Bubble Tea) — the program restores on normal return, so make every exit path normal: pass `tea.WithoutCatchPanics()` only if you install your own recovery, and always handle the error from `Run`:

```go
p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
if _, err := p.Run(); err != nil {
    fmt.Fprintln(os.Stderr, "fatal:", err) // after Run returns, the terminal is restored
    os.Exit(1)
}
```

Never call `os.Exit` from inside `Update`/`View`, and never `panic` in `View` — it fires with the alt screen active.

## Signals

| Signal | Required behaviour |
|--------|--------------------|
| `SIGWINCH` | treat size as live state; re-clamp scroll offsets, selection, and pane ratios |
| `SIGINT` / `SIGTERM` | restore terminal, flush persistence, exit non-zero-free (no half-written config) |
| `SIGHUP` | same as TERM; do not block waiting on network |
| `SIGTSTP` (`Ctrl-Z`) | leave alt screen + raw mode **before** stopping, or the shell inherits a broken terminal |
| `SIGCONT` | re-enter alt screen/raw mode, re-query size, force a full redraw |

Bubble Tea handles suspend/resume for `Ctrl-Z`; if you intercept `Ctrl-Z` yourself, you own the teardown/rebuild. In Crossterm, do the leave/enter explicitly around the stop.

## Clipboard

`OSC 52` is the only copy path that works over SSH and inside tmux without a helper binary:

```go
// Copy writes text to the terminal's clipboard via OSC 52.
// tmux requires `set -g set-clipboard on` to forward it.
func Copy(w io.Writer, text string) {
    b64 := base64.StdEncoding.EncodeToString([]byte(text))
    fmt.Fprintf(w, "\x1b]52;c;%s\x07", b64)
}
```

Fall back to `wl-copy` / `pbcopy` / `xclip -selection clipboard` when OSC 52 is known-unsupported. Always confirm the copy in the UI ("Copied address"), because the clipboard is invisible.

Large payloads are dropped by many terminals; cap OSC 52 at a few KB and use a temp file + message for anything bigger.

## Handing the Terminal to Another Program

For `$EDITOR`, `$PAGER`, `less`, `git diff`, or a shell: release the terminal, run the child inheriting stdio, then reclaim.

Bubble Tea has this built in — use it instead of `exec.Command` in a goroutine:

```go
c := exec.Command(editorFromEnv(), path) // $VISUAL, then $EDITOR, then a sane default
return m, tea.ExecProcess(c, func(err error) tea.Msg { return editorFinishedMsg{err: err} })
```

Rust: `disable_raw_mode` + `LeaveAlternateScreen`, spawn with inherited stdio, `wait`, then re-enter and force a redraw.

Never launch an interactive child while the alt screen is up.

## Logging and Debugging

`stdout` belongs to the renderer. A stray `println!`/`fmt.Println` corrupts the frame.

- Go: `f, _ := tea.LogToFile("debug.log", "tui"); defer f.Close()`, then `log.Printf`.
- Rust: `tracing_subscriber` with a `tracing_appender` file writer; never the default stderr writer while the alt screen is active.
- Gate it behind `--debug` or `TUI_DEBUG=1`; ship with file logging off by default.
- Write panics/backtraces to the log file **and** print a short human line after restore, so the user sees why it exited.
- `tail -f debug.log` in a second pane is the normal debugging loop.

## Startup

- Render the first frame from empty state immediately; never block startup on network, Docker, or disk scans. Show `Loading…` per pane.
- Query terminal size before the first render, and handle `0x0` (some CI/pipes report it) without dividing by zero.
- If you query the terminal (OSC 11 background, capability probes), always use a short timeout — many terminals never answer.
