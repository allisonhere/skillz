# TUI Interaction Patterns

## Navigation Contract

Make navigation predictable across the application.

Lists should generally accept arrows plus `j/k`. Spatial pane movement may use arrows plus `h/l` when it does not interfere with text editing. `Tab` and `Shift-Tab` are reliable focus-cycle fallbacks.

## Context Matters

Do not route global keys through active text inputs if those keys are valid editing characters. In forms, prioritize text editing and use explicit escape/back behavior.

When the cursor is at a boundary, spatial navigation can be useful if it is documented and predictable—for example, left at the beginning of a field returning focus to a preceding pane.

## Command Palette

Use a command palette when commands are numerous enough that permanent shortcuts become hard to discover.

Strong behavior:

- `Ctrl-K` opens when available
- typing fuzzy-filters commands
- first strong match is selected
- arrows or `j/k` navigate results
- `Enter` executes
- `Esc` closes without action
- rows show command name and shortcut when one exists
- commands may include contextual availability and disabled reasons

Prefer action names like:

- Restart container
- Open logs
- Copy address
- Change theme
- Refresh all

Avoid internal function names.

### Filtering

Subsequence (not substring) matching, scored so that word-start and prefix matches win, with stable ordering for ties:

```go
type Command struct {
    ID, Name, Shortcut string
    Enabled            bool
    DisabledReason     string
}

// scoreMatch returns (score, true) when every rune of query appears in name in
// order. Higher is better. Case-insensitive.
func scoreMatch(name, query string) (int, bool) {
    if query == "" {
        return 0, true
    }
    n, q := strings.ToLower(name), strings.ToLower(query)
    score, qi, prevMatch := 0, 0, false
    for i := 0; i < len(n) && qi < len(q); i++ {
        if n[i] != q[qi] {
            prevMatch = false
            continue
        }
        score += 1
        if i == 0 || n[i-1] == ' ' || n[i-1] == '-' {
            score += 8 // word start
        }
        if prevMatch {
            score += 4 // consecutive run
        }
        prevMatch = true
        qi++
    }
    if qi < len(q) {
        return 0, false
    }
    return score, true
}

// Filter keeps matching commands, best first, preserving input order on ties.
func Filter(cmds []Command, query string) []Command {
    type scored struct {
        cmd Command
        s   int
        i   int
    }
    var out []scored
    for i, c := range cmds {
        if s, ok := scoreMatch(c.Name, query); ok {
            out = append(out, scored{c, s, i})
        }
    }
    sort.SliceStable(out, func(a, b int) bool {
        if out[a].s != out[b].s {
            return out[a].s > out[b].s
        }
        return out[a].i < out[b].i
    })
    res := make([]Command, 0, len(out))
    for _, s := range out {
        res = append(res, s.cmd)
    }
    return res
}
```

Build the command list **from current context** each time the palette opens, so availability and disabled reasons are always true. Show disabled commands with their reason rather than hiding them — hiding teaches nothing.

## Status Hints

Show a small stable set of high-value commands on the main surface. Make hints context-sensitive in modals or special modes.

Do not duplicate the entire help screen in the footer.

## Help

`?` should open contextual help when practical. Group shortcuts by task rather than dumping one alphabetized list.

## Search and Filter

`/` is a strong search/filter convention. Make the mode obvious, provide a visible query, and make clearing/cancel behavior consistent.

For ranked search, show enough source context to explain results.

## Forms

Forms need explicit focus order, clear active field treatment, validation messages close to the relevant field, and safe handling of secrets.

Never reveal stored passwords/API keys merely because a field receives focus. Mask secrets by default.

## Confirmation

Confirmation dialogs should name the object and consequence:

`Delete feed “Example”? This removes its local articles.`

Prefer explicit verbs over generic `OK` where the action is consequential.

## Mouse

Mouse support is additive. Keyboard operation must remain complete.

Useful mouse behaviors:

- wheel scrolls the pane under pointer or focused pane consistently
- rows can be selected/clicked
- tabs can be clicked
- resize handles only when visually discoverable

Avoid hidden mouse-only capabilities.

## Quit Behavior

For applications with state that may be lost, optional quit confirmation is reasonable. For stateless tools, do not slow down quitting unnecessarily.

## Scrolling

Provide, in every scrollable pane:

- `j`/`k` and `↓`/`↑` — one row
- `Ctrl-D`/`Ctrl-U` — half page; `PgDn`/`PgUp` — full page
- `g`/`Home` — top; `G`/`End` — bottom
- keep the selected row visible with a small margin (2 rows) rather than pinning it to the edge
- show a scroll indicator when content is clipped (a rail, or `12–34/512`); a pane that silently hides content reads as a bug

For streaming content (logs, events, output): follow the tail while the view is at the bottom, stop following the moment the user scrolls up, resume at `G`/`End`, and show the state (`following` / `paused`).

Preserve the anchor across resize: the selected row stays selected and visible; do not reset to the top.

## Terminal Key Realities

The terminal, not your app, decides what some keys mean:

- `Ctrl-I` ≡ `Tab`, `Ctrl-M` ≡ `Enter`, `Ctrl-J` ≡ `\n`, `Ctrl-H` ≡ `Backspace` on many setups — binding them as distinct actions will surprise users.
- `Ctrl-S`/`Ctrl-Q` may be swallowed by flow control; `Ctrl-C` and `Ctrl-Z` carry strong shell expectations. Do not repurpose them without an obvious, documented reason.
- A bare `Esc` is indistinguishable from the start of an Alt/escape sequence without a timeout (~25–50 ms) or the kitty keyboard protocol. Never put a destructive action on `Esc`.
- Modified keys beyond `Ctrl-<letter>` (`Ctrl-Shift-K`, `Ctrl-Enter`) are unreliable in plain terminals. Fine as an accelerator, never as the only path.
- Mouse reporting and bracketed paste change how text arrives; a "paste" is a burst of keystrokes unless bracketed paste is on — guard against interpreting pasted characters as commands.

For multi-key sequences (leader keys), show the pending prefix in the status bar and time out (~1 s) back to normal.
