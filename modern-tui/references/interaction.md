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
