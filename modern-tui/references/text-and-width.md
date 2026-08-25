# Text, Width, and Truncation

## The Three Lengths

A terminal cares about **display width**, not bytes and not runes:

| Measure | Go | Rust | `"héllo→🇯🇵"` |
|---------|-----|------|--------------|
| bytes | `len(s)` | `s.len()` | wrong, always |
| runes / chars | `len([]rune(s))` | `s.chars().count()` | wrong for wide + combining |
| grapheme clusters | `uniseg.GraphemeClusterCount(s)` | `s.graphemes(true).count()` | cursor/edit unit |
| **display width** | `runewidth.StringWidth(s)` / `lipgloss.Width(s)` | `UnicodeWidthStr::width(s)` | **layout unit** |

Rules:

- Measure layout with display width. Never `len()`.
- Move a cursor by grapheme cluster, never by rune (combining marks and ZWJ emoji are multi-rune, one cell-group).
- `lipgloss.Width` ignores ANSI escapes; `runewidth.StringWidth` does not. If the string may already be styled, use the ANSI-aware measure.

## Never Slice a Styled String

Byte- or rune-slicing a string that contains escape sequences can cut mid-escape, leaking `\x1b[38;5` into the screen and corrupting every following cell. Truncate **before** styling, or use an ANSI-aware truncator.

Go — measure and truncate with an ellipsis that itself costs width:

```go
import (
    "strings"

    "github.com/charmbracelet/lipgloss"
    "github.com/mattn/go-runewidth"
)

// TruncateCells shortens plain (unstyled) text to at most w display cells,
// appending an ellipsis when it had to cut. Returns "" for w <= 0.
func TruncateCells(s string, w int) string {
    if w <= 0 {
        return ""
    }
    if runewidth.StringWidth(s) <= w {
        return s
    }
    if w == 1 {
        return "…"
    }
    return runewidth.Truncate(s, w, "…")
}

// PadCells right-pads plain text to exactly w display cells, truncating first.
func PadCells(s string, w int) string {
    s = TruncateCells(s, w)
    if gap := w - runewidth.StringWidth(s); gap > 0 {
        s += strings.Repeat(" ", gap)
    }
    return s
}

// FitStyled truncates already-styled content safely by measuring with the
// ANSI-aware width and letting lipgloss re-wrap inside a fixed box.
func FitStyled(styled string, w int) string {
    if lipgloss.Width(styled) <= w {
        return styled
    }
    return lipgloss.NewStyle().MaxWidth(w).Render(styled)
}
```

Rust:

```rust
use unicode_segmentation::UnicodeSegmentation;
use unicode_width::UnicodeWidthStr;

/// Shorten `s` to at most `w` display cells, appending `…` when cut.
pub fn truncate_cells(s: &str, w: usize) -> String {
    if w == 0 { return String::new(); }
    if s.width() <= w { return s.to_string(); }
    if w == 1 { return "…".to_string(); }
    let mut out = String::new();
    let mut used = 0usize;
    for g in s.graphemes(true) {
        let gw = g.width();
        if used + gw > w - 1 { break; }
        out.push_str(g);
        used += gw;
    }
    out.push('…');
    out
}
```

In Ratatui, prefer letting the framework clip: build `Line`/`Span`s and render into a computed `Rect`. Reach for manual truncation only for text you assemble yourself (headers, status segments, table cells).

## Two-Column Rows Without Drift

A row with a left label and a right suffix must reserve the suffix first, then fit the label:

```go
// Row renders "left … right" in exactly w cells; right is never truncated
// unless it alone exceeds w.
func Row(left, right string, w int) string {
    rw := runewidth.StringWidth(right)
    if rw >= w {
        return TruncateCells(right, w)
    }
    left = TruncateCells(left, w-rw-1)
    gap := w - runewidth.StringWidth(left) - rw
    return left + strings.Repeat(" ", gap) + right
}
```

This is the invariant tideui enforces with tests named after it: a long suffix must not push a row past its allocated width, and a long left side must not evict the suffix.

## Ambiguous and Zero Width

- East-Asian **ambiguous** characters (many box-drawing and arrow glyphs) are 1 cell in most terminals and 2 in CJK locales. `runewidth` decides from the locale env; force it deliberately (`runewidth.DefaultCondition.EastAsianWidth = false`) rather than inheriting surprise behaviour on a user's `LANG`.
- Emoji presentation is a coin flip across terminals. Prefer text-presentation glyphs (`▸ ● ◆ ✓ ✕ …`) in aligned columns; keep emoji out of columns whose alignment matters.
- Zero-width joiner sequences (flags, family/profession emoji) report widths that terminals disagree about. If a glyph must sit in a table, budget 2 cells and verify in at least two terminals.
- Strip or reject control characters and lone `\r` from any external text (log lines, container names, email subjects) before rendering — they move the cursor.

## Wrapping

- Wrap on grapheme boundaries, prefer word boundaries, and hard-break tokens longer than the pane.
- Wrapping is layout, so it belongs in the view; do not pre-wrap stored data to a width captured at load time — it will be wrong after the first resize.
