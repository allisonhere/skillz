---
name: omarchy-noctalia-plugin-scaffold
description: Scaffold and workflow conventions for the user's Omarchy/Noctalia Linux desktop plugin projects (launcher-installer, omarchy-noctalia, tablet, screen-widget) — Python-based tools that plug into the Noctalia shell / Omarchy Hyprland setup, often with a systemd service and/or a Quickshell/Noctalia plugin.toml. Use this whenever creating a new Linux desktop utility, Hyprland/Noctalia plugin, or system tray/widget tool for the user, or editing one of the existing ones.
---

# Omarchy/Noctalia plugin scaffold

The user runs Omarchy (Hyprland-based) with the Noctalia shell, and builds small
Python tools that integrate with it. `launcher-installer`, `omarchy-noctalia`,
`tablet`, and `screen-widget` all follow the same rough shape.

## Project layout

```
pyproject.toml       # setuptools, requires-python >=3.11, [project.scripts] entry point
src/<pkg>/            # the actual package
tests/                # pytest, one test module per source module
plugin/plugin.toml    # Noctalia plugin manifest, when the tool is a Noctalia plugin
```

`pyproject.toml` uses `setuptools>=68` as the build backend and declares a console
entry point under `[project.scripts]` — that's how these tools become a runnable
command after install, not via a hand-rolled shebang script.

## Noctalia plugin manifest

If the tool plugs into the Noctalia shell (as opposed to being a standalone CLI),
it needs a `plugin.toml` (sometimes under a `plugin/` or `noctalia/` subdirectory,
per `tablet`'s layout, which keeps the Noctalia integration separate from the core
`src/tabletctl.py` logic). Keep the plugin manifest thin — logic belongs in `src/`,
the manifest just wires it into the shell.

## System integration

Tools that run persistently (`tablet`) install a systemd **user** service via
`install.sh`, deployed from a `deploy/` or `systemd/` directory containing the unit
file — not installed as a system-wide service unless there's a specific reason to run
as root. Check for an existing `install.sh` before writing a fresh one; the pattern
(copy config, install unit, `systemctl --user enable`) is already established.

## Testing

`pytest`, one test module per source module (`test_cli.py` for `cli.py`,
`test_manager.py` for `manager.py`, etc.). Keep that mapping when adding new source
files — it's what the existing repos do and it makes it obvious what's covered.

## Starting a new plugin

When asked to build a new Omarchy/Noctalia tool, default to this scaffold rather than
a bare script:
1. `pyproject.toml` with a `[project.scripts]` entry point.
2. `src/<pkg>/` for logic, `tests/` mirroring it.
3. `plugin/plugin.toml` only if it actually integrates with the Noctalia shell UI —
   a pure background daemon or CLI utility (like `tablet`'s core) doesn't need one.
4. `install.sh` (+ a systemd user unit under `systemd/` or `deploy/`) if it needs to
   run persistently in the background.
