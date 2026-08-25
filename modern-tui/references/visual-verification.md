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
