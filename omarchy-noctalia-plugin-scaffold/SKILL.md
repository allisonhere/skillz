---
name: omarchy-noctalia-plugin-scaffold
description: Use when creating or editing an Omarchy/Noctalia desktop tool or Hyprland widget (launcher-installer, tablet, screen-widget) — Python core plus a Luau Noctalia plugin, systemd user units, and the plugin.toml/translations contract.
version: 1.0.0
license: MIT
platforms: [linux]
metadata:
  hermes:
    tags: [omarchy, noctalia, hyprland, quickshell, luau, python, systemd, linux-desktop]
    category: devops
---

# Omarchy/Noctalia plugin scaffold

The user runs Omarchy (Hyprland-based) with the Noctalia shell. Two shapes exist — check
which one the task needs before scaffolding.

**A. Pure Noctalia plugin (no Python).** `screen-widget` is exactly this: `plugin.toml`
+ `widget.luau` + `translations/en.json`, nothing else. Bar widgets and indicators
belong here.

**B. Python tool with an optional Noctalia widget.** `tablet` is the reference:

```text
pyproject.toml              # setuptools>=68, requires-python >=3.11, [project.scripts]
Makefile
src/tabletctl.py            # the logic
tests/test_tabletctl.py     # pytest, one module per source module
noctalia/plugin.toml        # Noctalia manifest (note: noctalia/, not plugin/)
noctalia/widget.luau        # the shell-side widget, in Luau
systemd/*.service           # user units (z13-tablet-mode.service, z13-squeekboard.service)
layouts/*.yaml              # tool data
install.sh                  # copies config, installs units, systemctl --user enable
deploy/
```

`launcher-installer` and `omarchy-noctalia` follow the same B shape. Logic stays in
`src/`; the manifest and widget stay thin.

`pyproject.toml` uses `setuptools>=68` as the build backend and declares a console entry
point under `[project.scripts]` — that's how these tools become a runnable command after
install, not via a hand-rolled shebang script.

## Noctalia plugin manifest

`plugin.toml` is a real schema, not freeform. Verified shape (from `screen-widget`):

```toml
id = "allieb/omarchy-workspaces"      # namespaced id
name = "Workspaces"
version = "1.1.3"
plugin_api = 19                        # MUST match the installed Noctalia plugin API
author = "allieb"
license = "MIT"
icon = "layout-dashboard"              # Lucide-style icon name
description = "Five configurable workspace indicators for Hyprland."
tags = ["bar", "workspaces", "hyprland"]
dependencies = ["hyprctl"]             # external commands the widget shells out to

[[widget]]                             # REQUIRED: registers the widget with the shell
id = "workspaces"
entry = "widget.luau"

[[setting]]                            # optional — tablet's manifest has none
key = "icon_layout"
type = "select"
label_key = "settings.icon_layout.label"
description_key = "settings.icon_layout.description"
default = "pill"
options = [
  { value = "pill", label_key = "settings.icon_layout.pill" },
  { value = "dots", label_key = "settings.icon_layout.dots" },
]
```

Rules that follow:

- `plugin_api` is a hard compatibility gate. Read the installed shell's current value
  before inventing one — copy it from a working plugin
  (`grep plugin_api ~/.config/noctalia/plugins/*/plugin.toml`) rather than guessing.
- `[[widget]]` (`id` + `entry = "widget.luau"`) is what registers the widget. Omit it and
  the plugin loads with nothing on the bar. Both `screen-widget` and `tablet` declare
  exactly one.
- Every user-visible string is a `*_key` resolved from `translations/en.json`. Don't
  inline English in `plugin.toml`; add the key to the translations file in the same
  change or the UI shows a raw key.
- `dependencies` lists external commands, so declare `hyprctl`/`wpctl`/etc. instead of
  assuming they exist.
- Settings drive the widget; keep behaviour in `widget.luau` (or the Python core), not in
  the manifest.

Templates to copy: `templates/plugin.toml`, `templates/en.json`.

## System integration

Tools that run persistently (`tablet`) install a systemd **user** service via
`install.sh`, deployed from a `deploy/` or `systemd/` directory containing the unit
file — not installed as a system-wide service unless there's a specific reason to run
as root. Check for an existing `install.sh` before writing a fresh one; the pattern
(copy config, install unit, `systemctl --user enable`) is already established.

## Testing and lint

`pytest`, one test module per source module (`test_cli.py` for `cli.py`,
`test_manager.py` for `manager.py`). Keep that mapping when adding new source files —
it's what the existing repos do and it makes coverage obvious.

`tablet` also runs **ruff** and **mypy** (both `.ruff_cache/` and `.mypy_cache/` are
present, alongside a `Makefile`), so "tests pass" isn't the whole gate:

```sh
grep -E '^[a-z][a-z-]*:' Makefile | cut -d: -f1     # see what's wired
ruff check . && mypy src && pytest
```

## Noctalia verification

Luau widgets have no test harness here — verify them in the running shell.

Before editing a manifest, copy the installed API value from a known-working plugin:

```sh
grep -R '^plugin_api' ~/.config/noctalia/plugins/*/plugin.toml
```

After installing or editing, reload the shell/plugin using the repo's existing
`install.sh`, Noctalia command, or Hyprland/session restart pattern already present in
that project. Then inspect shell logs for raw translation keys, manifest parse errors,
or widget registration failures:

```sh
journalctl --user -xe | grep -i noctalia
grep -R 'settings\..*\.label\|settings\..*\.description' noctalia translations 2>/dev/null
```

Failure signatures to check first:

- raw `settings.*` text in UI: missing `translations/en.json` key
- plugin appears installed but no widget is available: missing or wrong `[[widget]]`
- plugin rejected at load: stale `plugin_api`
- widget silently blank: missing external command in `dependencies` or a failing
  `hyprctl`/`wpctl` shell-out

For Python-backed tools, verify both layers: `ruff`/`mypy`/`pytest` for the Python core,
then reload Noctalia and exercise the widget against the installed command.

## Starting a new plugin

1. Decide shape A (pure Luau widget) or B (Python tool, optional widget). A bar
   indicator is A; anything with state, a daemon, or a CLI surface is B.
2. Shape A: `plugin.toml` + `widget.luau` + `translations/en.json`. Done — no
   `pyproject.toml`.
3. Shape B: `pyproject.toml` with a `[project.scripts]` entry point, `src/<pkg>/`,
   `tests/` mirroring it, and a `Makefile` wiring lint/type/test.
4. Add `noctalia/plugin.toml` + `noctalia/widget.luau` only if it surfaces in the shell
   UI — a pure background daemon or CLI utility doesn't need one.
5. Add `install.sh` + `systemd/*.service` (user units, `systemctl --user enable`) only if
   it must run persistently. Check for an existing `install.sh` first; the pattern
   (copy config, install unit, enable) is already established.
