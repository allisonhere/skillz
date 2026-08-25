# Install modern-tui

This bundle follows the Agent Skills `SKILL.md` pattern used by Claude and Hermes.

## Hermes

Copy the folder into a Hermes skills category, for example:

```bash
mkdir -p ~/.hermes/skills/software-development
cp -a modern-tui ~/.hermes/skills/software-development/modern-tui
```

Start a new Hermes session, or use Hermes' current skill refresh/reset mechanism, then test with:

```text
/modern-tui review this TUI layout
```

## Claude Code

Use the current Claude Code skills/plugin mechanism. If your Claude Code setup scans a personal skills directory, install this folder there; otherwise place it under the `skills/modern-tui/` directory of a Claude Code plugin.

The required portable unit is the folder containing `SKILL.md`; `references/` and `templates/` are supporting resources.

## Shared Source Strategy

To avoid maintaining divergent copies, keep this bundle in one source-controlled directory and symlink/copy it into each agent's expected skills location. Hermes supports external skill directories as well, so an external shared directory can be preferable when configured.
