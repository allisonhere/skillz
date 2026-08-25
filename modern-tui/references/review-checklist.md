# Modern TUI Review Checklist

## Information Architecture

- Is the primary task obvious?
- Are panes/views based on real information relationships?
- Is important context visible without unnecessary navigation?
- Is there wasted vertical or horizontal space?

## Visual Hierarchy

- Is focused pane unmistakable?
- Is selected row unmistakable?
- Are focus and selection distinguishable?
- Are borders communicating something useful?
- Is metadata quieter than primary content?
- Is the interface over-colored?

## Interaction

- Do arrows work where users expect them?
- Are Vim keys optional rather than mandatory?
- Are global shortcuts consistent?
- Are text inputs protected from conflicting global bindings?
- Is `Esc` behavior predictable?
- Does `?` provide useful contextual help?
- Would a command palette reduce shortcut clutter?

## Status and Feedback

- Are loading states visible?
- Are empty states intentional?
- Are errors human-readable?
- Are destructive actions confirmed?
- Does background work leave input responsive?

## Layout and Resizing

- Does it work in a narrow tmux/zellij pane?
- Does it work at common 80x24 dimensions?
- Does it remain bounded at tiny dimensions?
- Does resizing preserve valid focus/selection?
- Are pane ratios sensible and bounded?

## Accessibility

- Is status encoded by more than color?
- Are focus and selection contrast sufficient?
- Is there an ASCII/reduced-Unicode path when needed?
- Are symbols width-safe?

## Architecture

- Is rendering mostly state-to-view?
- Are slow operations outside the render/input loop?
- Is service logic separated from UI components?
- Are reusable behaviors actually shared rather than copied?
- Are semantic actions testable separately from raw key events?

## Polish

- Are shortcut hints concise?
- Are modals proportionate to their task?
- Are long strings truncated/wrapped intentionally?
- Are labels consistent in capitalization and terminology?
- Do theme changes apply coherently to overlays and selection states?
- Are graphs compact and meaningful rather than decorative?
