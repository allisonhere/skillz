# Performance and Scale

## Budget

| Interaction | Budget |
|-------------|--------|
| keypress → visible change | < 50 ms (feels instant) |
| single frame render | < 16 ms typical, < 33 ms worst case |
| startup → first frame | < 100 ms, from empty state |

If a frame cannot meet that, the fix is less work per frame — not a spinner.

## Render Only On Change

- Redraw on state change, not on a timer. A 60 Hz repaint of a static screen burns CPU and battery for nothing.
- Metrics tick at **1–4 Hz**; sparkline history is data, not animation.
- Coalesce bursts: drop superseded resize events and collapse rapid updates for the same key into one pending state change.
- Debounce search/filter input (~50–120 ms) when the filter is expensive; never block the keystroke on it.
- Cache expensive derived strings (wrapped bodies, syntax-highlighted blocks, formatted tables) keyed by `(content, width, theme, density)` and invalidate on resize/theme change.

## Virtualize Long Lists

Formatting 100k rows to display 40 is the single most common TUI performance bug. Format only the visible window:

```go
// VisibleWindow returns the [start, end) row range to render, keeping the
// selected row inside a `scrolloff`-row margin. total is the row count,
// height the number of body rows the pane can draw.
func VisibleWindow(total, height, selected, offset, scrolloff int) (start, end, newOffset int) {
    if height <= 0 || total <= 0 {
        return 0, 0, 0
    }
    if scrolloff*2 >= height {
        scrolloff = 0 // pane too short for margins
    }
    selected = min(max(selected, 0), total-1)
    if selected < offset+scrolloff {
        offset = selected - scrolloff
    }
    if selected > offset+height-1-scrolloff {
        offset = selected - height + 1 + scrolloff
    }
    offset = min(max(offset, 0), max(0, total-height))
    start = offset
    end = min(offset+height, total)
    return start, end, offset
}
```

Then render `rows[start:end]` only. The same math drives the scrollbar (`offset/total`, `height/total`) with no extra state.

Rules that follow from it:

- Keep the **offset** in UI state and derive the window every frame; do not keep a slice of pre-rendered rows.
- Re-clamp offset and selection on resize and on data change (rows can vanish under you).
- Cap unbounded streams (logs, events, output) with a ring buffer and say so in the UI ("last 10 000 lines").
- For log/tail panes, keep a `follow` flag: on when pinned to the bottom, off as soon as the user scrolls up, back on at `G`/`End`.

## Keep Work Off the UI Loop

Anything that can block belongs in a task/command: HTTP, SSH, Docker/Podman, DB queries, directory walks, subprocesses, AI inference, update checks.

- Model each as explicit `started → succeeded | failed` messages and render those states.
- Make in-flight work cancelable (`Esc`), and cancel superseded requests (typing in a search box must not leave 12 requests racing).
- Never `await`/block inside a render function, and never hold a lock across a render.
- Bound concurrency (a small worker pool) so 200 containers don't open 200 sockets.

## Cheap Rendering Habits

- Build frames with a reused `strings.Builder` / preallocated `String`; avoid `+=` in row loops.
- Style objects are values — build them once per theme change, not per row.
- Avoid re-measuring the same string repeatedly; measure once and pass the width along.
- Prefer one pane-level styled block over per-cell styling when the result is identical — fewer escape sequences means less bytes on the wire, which matters over SSH.

## Measure, Don't Guess

- Go: benchmark the row/frame builders (`go test -bench . -benchmem ./internal/ui/`); tideui benchmarks its modal shadow blend this way.
- Rust: `cargo build --release` before judging speed (debug renders can be 10× slower); use `criterion` for hot helpers.
- Track worst-case frame time in `--debug` and log frames over budget with the state that produced them.
