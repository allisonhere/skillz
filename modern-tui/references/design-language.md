# Modern TUI Design Language

## Hierarchy First

Establish hierarchy with position, spacing, typography-like emphasis, muted metadata, and focus treatment before adding borders.

Use borders when they communicate containment, focus, or modal boundaries. A border is not automatically clutter: a consistent full-pane focus border can be excellent when it is part of the product language and has sufficient contrast.

## Density

Default toward compact density for developer tools and monitoring surfaces. Offer a comfortable mode when reading-heavy content benefits from breathing room.

Do not create large vertical gaps between list rows merely to imitate desktop/web UI.

## Multi-pane Patterns

### Sidebar + stacked right

Best when a primary collection controls two related views, for example feeds → articles + content.

### Three-column

Best when three levels of hierarchy need simultaneous visibility and horizontal space is available.

### Sidebar + main

Best for narrow screens or when the third pane is unnecessary.

### Tabbed

Best as a responsive fallback when multiple panes cannot remain useful simultaneously.

### Floating

Best for transient inspectors, quick actions, or secondary context—not permanent primary navigation.

## Responsive Collapse Order

When width or height becomes constrained:

1. remove decorative padding
2. shorten metadata
3. truncate secondary labels
4. switch to compact density
5. collapse secondary panes
6. use tabs for peer panes
7. show a minimum-size message only when meaningful operation is impossible

Never panic or render outside the terminal bounds.

## Pane Resizing

For adjustable panes, use bounded ratios with small predictable steps. Keep minimum widths that preserve useful content. If persisted display preferences exist, persist ratios.

## Selection and Focus

Selection and focus are different states. A selected row inside an unfocused pane should not look identical to the selected row in the focused pane.

Use more than hue alone: background, rail, glyph, inverse treatment, border, text weight, or another redundant cue.

## Soft Modal Language

Prefer centered, restrained overlays with:

- short embedded title
- concise content
- one obvious focus/selection cue
- quiet action hints
- enough surrounding context to understand what the modal affects

Avoid huge modal boxes that consume the entire screen for tiny tasks.

## Theme Preview

For theme pickers, preview changes live as the user navigates when rollback is reliable. Cancel should restore the previously confirmed theme. Confirm should persist only the selected theme.

## Theme Contrast

Treat contrast as a property to verify, not assume. Focus borders and selected backgrounds should remain visible across very dark, very light, and low-saturation themes.

## Terminal Background

Do not assume pure black. Respect the application theme and terminal background behavior. If emitting OSC background controls, keep emission under application ownership and reset cleanly on exit.

## Icons and Unicode

Use Unicode icons only where they improve scanability. Provide plain-text or ASCII alternatives when compatibility is a goal.

Do not let icon width ambiguity break alignment.

## Visual Character

Distinctive does not mean loud. Character can come from:

- a recognizable modal language
- a consistent focus rail or border treatment
- thoughtful status-line phrasing
- tasteful glyphs
- strong theme curation
- compact graphs
- motion used sparingly for meaningful state
