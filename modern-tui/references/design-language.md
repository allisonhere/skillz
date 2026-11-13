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

### Top-to-bottom light sweep

A restrained shine moving from the top of a modal to the bottom is a preferred
optional flourish for this user's Tide-style interfaces. Use it where it fits
the application's visual language; it is not a requirement for every modal.

- Move a soft highlight band downward across the border or surface. Derive its
  colors from the active theme and preserve text, selection, and focus contrast.
- Keep geometry and text stationary during the sweep. Do not simulate shine by
  moving content or changing which row is selected.
- A short single pass on opening is a useful default. Avoid a permanent repaint
  loop or repeated flashes while the user reads or types.
- With no color, reduced motion, or limited terminal capabilities, render the
  normal static modal. The effect must carry no essential information.

### Top-to-bottom collapse

A vertical closing transition is another preferred optional pattern. For a
literal top-to-bottom collapse, keep the bottom edge anchored and move the top
edge down until the panel disappears; a centered squeeze is a different effect.

- Render the modal at its normal geometry, then clip its visible rows as the top
  edge advances. Avoid repeatedly reflowing text into a smaller box.
- Restore the underlying application cells as each row is uncovered, including
  any shadow. Never leave border or shadow fragments behind.
- Closing must remain immediate logically: release the modal's input ownership,
  restore the intended focus, and treat the remaining animation as noninteractive.
  A new modal or resize may finish the old transition immediately.
- Reduced motion uses an immediate close. If the host already has an animation
  duration setting, honor it instead of adding a second timing system.

### Implementing modal motion

Use the host's existing animation mechanism, inspecting TideUI's actual API
before adding an application-owned effect. These patterns do not imply that
TideUI exposes a ready-made shine or collapse widget.

In Bubble Tea, advance elapsed-time progress through bounded tick messages;
keep rendering pure and never sleep in Update or View. Tag ticks with the modal
session so a late tick cannot animate a newly opened modal. Stop scheduling when
the pass finishes, the modal closes, or the application exits. Shine and collapse
may be used independently or together; neither should delay keyboard actions.

Verify opening, dismissal, rapid close/reopen, resizing mid-transition, and
light/dark/high-contrast themes. Test first, intermediate, and final frames with
a deterministic clock, bounded output at tiny sizes, restoration of exposed
background cells, and the static reduced-motion path.

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

## Before / After

**Footer** — every command vs. the five that matter:

```text
✗  q:quit ?:help /:search r:refresh R:restart s:stop S:start d:delete l:logs i:inspect e:exec p:prune t:theme m:manage c:copy Tab:next
✓  m manage   / search   l logs   ? help
```

**List row** — decoration vs. hierarchy (primary bright, metadata quiet, state as glyph + color):

```text
✗  ┌────────────────────────────────────────┐
   │ [RUNNING] postgres-main   (5432)  98%  │
   └────────────────────────────────────────┘

✓  ● postgres-main      5432   cpu 12%  mem 98%  ▇▆▅▃▂
     up 4d
```

**Empty state** — blank vs. next action:

```text
✗  (nothing)
✓  No containers running — press n to create one, r to refresh
```

**Modal** — full-screen box for a yes/no vs. a proportionate soft panel:

```text
✗  a 78×22 bordered box containing one question
✓  ╭─ Delete feed ───────────────────────────╮
   │ Delete "Example Weekly"?                │
   │ Removes 412 local articles.             │
   │                                         │
   │        [ Delete ]   Cancel              │
   ╰─────────────────────────────────────────╯
     enter confirm   esc cancel
```

**Focus** — hue-only vs. redundant cue (rail + background + brighter text):

```text
✗  selected row differs only by a slightly bluer foreground
✓  ▌ selected row on a filled background
```
