# Capabilities and Degradation

Decide capabilities **once at startup**, store them in state, and render from that — never probe inside `View`.

## Color

Precedence, highest first:

1. `NO_COLOR` set (any value) → no color at all. Respect it; it is a user accessibility choice.
2. Explicit app flag/config (`--color=never|auto|always`).
3. `CLICOLOR_FORCE` non-zero → keep color even when piped.
4. stdout is not a TTY → no color, no alt screen, plain line output.
5. `TERM=dumb` or unset → no color, no cursor addressing.
6. `COLORTERM=truecolor|24bit` → 24-bit.
7. `TERM` contains `256color` → 256.
8. Otherwise → 16 colors.

Go: Lipgloss/termenv already implement this; get the profile once (`lipgloss.DefaultRenderer().ColorProfile()`), and for tests build a renderer with a forced profile instead of mutating globals.

Rust: Ratatui has no detector — read the env yourself into a `ColorSupport` enum and map theme colors through it (`Color::Rgb` → nearest xterm-256 → nearest ANSI-16).

Degrading 24-bit to 16 colors collapses distinct semantic roles onto the same slot. Verify **focus, selection, error** remain distinguishable at 16 colors, and add a redundant non-color cue (rail, glyph, bold, inverse) wherever they are not.

## Contrast Is Computed, Not Assumed

Themes are user data; a hand-picked "muted" color can vanish on some backgrounds. Compute relative luminance and raise the color until it clears a ratio floor:

```go
func srgbLinear(v float64) float64 {
    if v <= 0.04045 {
        return v / 12.92
    }
    return math.Pow((v+0.055)/1.055, 2.4)
}

// Luminance is WCAG relative luminance for an 8-bit sRGB triple.
func Luminance(r, g, b uint8) float64 {
    lr, lg, lb := srgbLinear(float64(r)/255), srgbLinear(float64(g)/255), srgbLinear(float64(b)/255)
    return 0.2126*lr + 0.7152*lg + 0.0722*lb
}

// ContrastRatio returns the WCAG ratio (1.0 … 21.0) between two colors.
func ContrastRatio(l1, l2 float64) float64 {
    hi, lo := math.Max(l1, l2), math.Min(l1, l2)
    return (hi + 0.05) / (lo + 0.05)
}
```

Useful floors: **≥ 4.5** body text, **≥ 3.0** selected-row background vs pane background and large/bold text, **≥ 7.0** focus indication you want unmistakable. Nudge lightness toward or away from the background until the floor is met rather than hardcoding a second palette. tideui does exactly this (`paneFocusMinContrast = 7.0`, `selectedBgMinContrast = 3.0`) and asserts it in tests — copy the approach, not the constants.

`isDark(bg)` via `Luminance < 0.179` is a sound switch for choosing light/dark-appropriate derived colors.

## Unicode Level

Define levels and pick one at startup:

| Level | Use |
|-------|-----|
| `ascii` | `+-\|` borders, `[x]`, `->`, `#` meters — guaranteed everywhere, plus screen readers and `TERM=dumb` |
| `basic` | box drawing `│ ─ ╭ ╮`, block elements `█ ▓ ░ ▁▂▃`, arrows — safe on any modern terminal |
| `full` | powerline/nerd glyphs, emoji — only when the user opts in |

Never auto-detect nerd fonts; make them opt-in config. Offer `--ascii` and a per-theme flag (tideui models this as a theme property: its `vt52` theme reports `UsesASCII()`), so one switch changes borders, meters, and status glyphs together.

## Non-Interactive Output

A TUI that is piped or run in CI should degrade, not hang:

- stdout not a TTY → print a plain, line-oriented summary and exit; do not enter the alt screen.
- Provide an explicit `--plain` / `--once` mode for scripts and screen readers.
- stdin not a TTY → do not enable raw mode or mouse reporting.

## Multiplexers and Remote

- tmux/zellij: assume 16/256 color unless `COLORTERM` says otherwise; OSC sequences may need tmux passthrough (`\x1bPtmux;` + escaped payload + `\x1b\\`) and `set-clipboard on` for OSC 52.
- Layout must survive a narrow split — test at 80×24 and at 40 columns.
- Over SSH, keystroke latency is real: never require rapid multi-key timing, and keep the frame small enough that a redraw is not visibly slow.
- Kitty keyboard protocol (or `modifyOtherKeys`) is what makes `Ctrl-Shift-…`, real `Esc`, and key-release reliable. Enable when available, always keep a plain-terminal binding for the same action, and disable it on exit.
