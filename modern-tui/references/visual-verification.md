# Visual Verification

Use visual checks when layout, styling, focus, truncation, or animation changed. Tests
catch invariants; screenshots catch whether the interface still feels and reads right.

## Capture Frames

Prefer deterministic terminal sizes and color settings:

```bash
NO_COLOR=1 COLUMNS=80 LINES=24 go test ./internal/ui -run TestSnapshot -v
```

For Bubble Tea programs, use `github.com/charmbracelet/x/exp/teatest` for one or two
flows that need a real event loop. Send keys, wait for stable output, and assert the
final frame or a meaningful substring. Keep exact-frame goldens for core surfaces only.

For Ratatui, use `TestBackend` and snapshot the buffer with `insta` when exact layout is
part of the product. Review the snapshot diff before accepting it.

## Manual Terminal Pass

Run the app in the smallest realistic surfaces before calling UI work done:

```bash
resize -s 24 80 2>/dev/null || true
NO_COLOR=1 ./app
TERM=dumb ./app
```

Also check a narrow split around 40 columns, a normal 80x24 terminal, and one larger
desktop-sized terminal. In each size, verify focused pane, selected row, status hints,
modal placement, long text, empty state, and error state.

## Driving the real app with keys

A check that needs input needs a pty you own, not a pipe. `script -qec "stty cols 120 rows 40; exec ./app" capture.log`
gives the app a real terminal, but it does not reliably deliver piped keystrokes; fork the pty yourself, set the
window size (`TIOCSWINSZ`, including the pixel fields when the app's geometry depends on the cell shape), write the
keys with pauses between them, and read until the app goes quiet.

A TUI interrogates its terminal before the first frame — background colour (`OSC 11`), cursor position (`CSI 6n`) —
and waits politely forever if nothing answers. A bare pty has to answer both, or the capture is a few dozen bytes of
query that reads exactly like a crash.

The capture is a diff stream, not a series of screens: the renderer redraws only the cells that changed, with
absolute cursor moves. Any "read the last frame" text heuristic therefore returns stale or partial content, and a
half-drawn frame will happily place a row beside content it never shared a screen with. Replay the bytes into a grid
— a small emulator handling CUP/EL/ED, the cursor moves and SGR is enough — and, while calibrating, print the screen
after **every** key. That is how you see a pane answer a keypress one press late; capturing per key and replaying
each capture through the final-screen reader is the cheap version.

## Driving the app with tmux

When you only need to press keys and read the screen, tmux is simpler than owning a pty: it
delivers keystrokes, answers the terminal's startup queries, and holds a real grid.

```bash
tmux new-session -d -s shot -x 110 -y 30 \
  "env COLORTERM=truecolor XDG_DATA_HOME=$S/data XDG_CONFIG_HOME=$S/config ./app"
sleep 2; tmux send-keys -t shot S; sleep 1; tmux send-keys -t shot Down
tmux capture-pane -p -t shot              # the visible text
tmux capture-pane -e -p -t shot | grep -o '48;2;[0-9]*;[0-9]*;[0-9]*' | sort | uniq -c
tmux kill-session -t shot
```

`-e` keeps the color escapes, which is how you check that a shadow is darker than the page or a
focus rail moved. Run it against a **copy** of the real data and config (`cp` the database and
`config.toml` into a scratch directory and point `XDG_DATA_HOME` / `XDG_CONFIG_HOME` at it), with
update checks and background refresh switched off in the copied config, so a capture can never
change the user's library. Capture at two or three sizes; a toggle that sat below the fold at
34 rows was invisible in the reporter's window.

## What To Look For

- no line exceeds the terminal width
- no content appears below the allocated height
- selected and focused states remain distinguishable without color
- status/footer keeps the most important left-side text when hints are clipped
- modal borders and content stay inside tiny viewports
- CJK, emoji, combining marks, and long unbroken strings do not corrupt alignment
- loading/error/empty states are visible instead of blank panes
- animation or streaming output does not make keyboard input lag

## Screenshot Artifacts

If you save screenshots or frame dumps, keep them under an existing test/snapshot
location such as `testdata/`. Do not commit ad hoc terminal captures unless the repo
already treats them as goldens.

Keep captures out of the repo root, too: a `pic.png` screenshot there was committed twice by a
release script's `git add -A`, once under the message "chore: release v1.2.1". Put scratch
captures in a temp directory.
