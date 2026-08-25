# Testing a TUI

Three layers, cheapest first.

## 1. State transitions (most of your tests)

Test `key → action → state`, with no rendering involved. Fast, stable, and where the real logic lives.

```go
func TestJMovesSelectionAndClampsAtEnd(t *testing.T) {
    m := Model{items: make([]Item, 3), sel: 2, h: 10}
    next, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
    m = next.(Model) // handleKey returns tea.Model; assert back to the concrete type
    if m.sel != 2 {
        t.Fatalf("selection must clamp at the last row, got %d", m.sel)
    }
}
```

Cover: focus traversal, selection clamping at both ends, `Esc` at every overlay depth, filter narrowing then clearing, destructive action requiring confirmation, and async `started/succeeded/failed` transitions.

Route keys through a semantic action layer and you can test the actions directly, without synthesizing key events.

## 2. Bounded-render invariant (the highest-value test)

Rendering must never panic and never exceed its allocated box — at *any* size. Randomize the size; do not hand-pick three.

Go:

```go
func TestRenderNeverExceedsRequestedSize(t *testing.T) {
    for i := 0; i < 2000; i++ {
        w := 1 + rand.Intn(200)
        h := 1 + rand.Intn(60)
        m := newTestModel(w, h) // realistic content: long unicode strings, empty lists, errors
        out := m.View()         // must not panic
        lines := strings.Split(out, "\n")
        if len(lines) > h {
            t.Fatalf("%dx%d produced %d lines", w, h, len(lines))
        }
        for n, line := range lines {
            if got := lipgloss.Width(line); got > w {
                t.Fatalf("%dx%d line %d width %d", w, h, n, got)
            }
        }
    }
}
```

Rust, with Ratatui's test backend:

```rust
#[test]
fn render_never_panics_or_overflows() {
    for w in 1u16..120 {
        for h in 1u16..40 {
            let mut term = Terminal::new(TestBackend::new(w, h)).unwrap();
            let mut app = App::sample();
            term.draw(|f| render(f, &mut app)).unwrap(); // panics fail the test
            assert_eq!(term.backend().buffer().area.width, w);
        }
    }
}
```

Feed it hostile content: CJK, emoji, combining marks, a 10 000-char line with no spaces, empty collections, and an error state. This single test is what makes "never render outside the terminal bounds" real instead of aspirational.

## 3. Golden / snapshot frames (a few, deliberately)

Snapshot only the surfaces whose exact look is part of the product: main layout, focused vs unfocused pane, selected row, one modal, one empty state, one error state.

- Force a deterministic color profile and fixed size, or the file churns per machine. Go: build a `lipgloss.NewRenderer(io.Discard)` with an explicit profile. Rust: `TestBackend` is already deterministic; `insta` gives review tooling.
- Write goldens under `testdata/`, regenerate with an explicit flag (`go test ./internal/ui -update`), and **read the diff** before accepting it.
- Never snapshot a screen containing time, durations, IDs, or paths without normalizing them first.
- Go end-to-end: `github.com/charmbracelet/x/exp/teatest` drives a real program (send keys, wait for output, assert final frame) — use it for one or two flows, not for everything.

## Assertions Worth Writing

- focused and unfocused panes render **differently** (catches theme regressions that erase focus)
- selected row differs from unselected in more than color (a glyph or background, not hue alone)
- selection stays visible after resize and after the data set shrinks
- a long single-line value is truncated with an ellipsis, not wrapped, in a one-line field
- the status bar keeps its left segment when the right hints do not fit
- modal geometry stays inside the viewport at 20×6
- theme contrast: every built-in theme clears the ratio floors (loop over the theme list and assert)
- ASCII level emits no non-ASCII bytes: `for _, r := range out { if r > unicode.MaxASCII { t.Fatal(...) } }`

## Commands

```bash
go test ./... -race                          # all tests
go test ./internal/ui -run TestRender -v     # one area
go test ./internal/ui -update                # regenerate goldens (then read the diff)
go test -bench . -benchmem ./internal/ui     # frame/row builder cost

cargo test                                   # all tests
cargo test render_never -- --nocapture       # one test
cargo insta review                           # accept/reject snapshot changes
```

## Manual Pass Before Calling It Done

Resize to 40 columns and back; run inside a tmux split; run at 80×24; `Ctrl-Z` then `fg`; pipe stdout to `cat`; run with `NO_COLOR=1`; switch to the lightest and darkest themes; open and cancel every modal; trigger an error path (stop the daemon, break the network).
