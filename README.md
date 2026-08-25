# skillz

Personal repository of reusable AI agent skills, written once and installed into
whichever local coding agents support them.

## Layout

```text
skillz/
├── install-skill        # installer script (see below)
├── README.md
├── modern-tui/
│   ├── SKILL.md          # required: name + description frontmatter
│   ├── references/
│   └── templates/
└── <another-skill>/
    ├── SKILL.md
    └── ...
```

Each top-level directory containing a `SKILL.md` is a skill. There's no registry
file to maintain — the installer discovers skills by scanning for `*/SKILL.md`.

## Adding a new skill

Create a new top-level directory with a `SKILL.md` (YAML frontmatter with at least
`name` and `description`, then the skill body). Anything else in the directory
(`references/`, `templates/`, scripts, etc.) is carried along automatically —
the whole directory is the installable unit.

```bash
mkdir another-skill
$EDITOR another-skill/SKILL.md
```

## Installing skills

Run `install-skill` from anywhere — it resolves its own location, so it doesn't
matter what your current directory is.

```bash
# Install one skill to all default targets (agents, codex, claude)
~/Projects/skillz/install-skill modern-tui

# Install several at once
~/Projects/skillz/install-skill modern-tui another-skill

# Install to a specific agent only
~/Projects/skillz/install-skill modern-tui --target codex

# Explicit "everything" (same set as the default, spelled out)
~/Projects/skillz/install-skill modern-tui --all

# List what's in the repo and where each skill is currently installed
~/Projects/skillz/install-skill --list

# Help
~/Projects/skillz/install-skill --help
```

Adding `~/Projects/skillz` to your `PATH` (or symlinking `install-skill` onto
your `PATH`) lets you drop the leading path and just run `install-skill modern-tui`.

## How installs behave

`~/Projects/skillz` is the canonical, editable source. Installed skills are
**symlinks** back into this repo wherever the target agent just does a plain
directory scan for `*/SKILL.md` — so an edit here shows up for every agent
immediately, with nothing to re-run.

| Target    | Destination                              | Method  | Notes |
|-----------|-------------------------------------------|---------|-------|
| `agents`  | `~/.agents/skills/<skill>`                | symlink | shared/generic skills directory used across tools |
| `codex`   | `~/.codex/skills/<skill>` (`$CODEX_HOME`) | symlink | Codex's own `skill-installer` does a plain `os.listdir` + `isdir` scan, which follows symlinks |
| `claude`  | `~/.claude/skills/<skill>`                | symlink | Claude Code personal skills directory |
| `hermes`  | `~/.hermes/skills/<category>/<skill>`     | copy    | opt-in only, see below |

The installer never deletes anything. If a destination already exists and
isn't already a correct symlink into this repo, it leaves it alone and tells
you to pass `--force` — which moves the old path aside to
`<dest>.bak.<timestamp>` before installing, rather than overwriting it.

Re-running `install-skill` on an already-installed skill is a no-op (it
detects the existing correct symlink and reports "already installed").

### Hermes is opt-in and copy-based, not symlinked

Hermes' `~/.hermes/skills/` tree is actively managed by a curator process
(`.curator_ledger.jsonl`, `.bundled_manifest`, `.usage.json`) and organized
into categories (e.g. `software-development/`). Because of that, the
installer:

- only touches it when you explicitly pass `--target hermes` (it's excluded
  from `--all` and from the no-flags default),
- **copies** the skill directory into
  `~/.hermes/skills/<category>/<skill>` instead of symlinking, using the
  `metadata.hermes.category` field from the skill's `SKILL.md` frontmatter
  (falls back to `general` if absent),
- never edits `~/.hermes/config.yaml` — enabling Hermes' external skill
  directories feature (so it could read `~/.agents/skills` directly) is a
  manual config change left for you to make if you want it.

Because it's a copy, re-editing the skill in this repo won't update the
Hermes copy automatically — re-run `install-skill <skill> --target hermes`
after changes, or verify recognition with Hermes' own skill list/refresh
command.

## Updating the repository

```bash
cd ~/Projects/skillz
git pull
```

Since installs are symlinks, a `git pull` that changes an existing skill's
files is immediately live for every symlinked target with no reinstall step.
A `git pull` that adds a brand-new skill directory still needs one
`install-skill <new-skill>` to create its symlinks.

## Supported agents / targets

- **Codex** — confirmed: its skill directory is a flat `$CODEX_HOME/skills/<name>/SKILL.md`
  scan that follows symlinks.
- **Claude Code** — confirmed: personal skills load from `~/.claude/skills/<name>/SKILL.md`;
  Claude Code picked up the symlinked `modern-tui` immediately after install.
- **`~/.agents/skills`** — a shared skills directory already used as an external
  skill source by other tools on this machine (e.g. referenced in Hermes' own
  config as an example external directory); installed the same way as Codex/Claude.
- **Hermes** — supported but opt-in/copy-based; see above.
