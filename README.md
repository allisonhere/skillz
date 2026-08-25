# skillz

My personal collection of AI agent skills — I write each one once here and install
it into whichever local coding agents can use it (Claude Code, Codex, Hermes, and
anything else that shows up later).

## Layout

```text
skillz/
├── install-skill        # installer script (see below)
├── check-skill           # validator script (see below)
├── tui/                   # the interactive picker's source (see below)
├── README.md
├── modern-tui/
│   ├── SKILL.md          # required: name + description frontmatter
│   ├── references/
│   └── templates/
└── <another-skill>/
    ├── SKILL.md
    └── ...
```

A skill is just any top-level directory with a `SKILL.md` in it. There's no
registry to keep in sync — the installer finds skills by scanning for `*/SKILL.md`.

## What's in here

| Skill | Use it when |
|-------|-------------|
| `modern-tui` | designing, building, or reviewing any keyboard-first terminal UI |
| `tide-tui-ecosystem` | editing a Tide-family repo — cross-repo duplication, `tideui`, release shape |
| `go-cli-workflow` | building, testing, versioning, or releasing a Go CLI |
| `astro-content-seo` | adding content or deploying an Astro content site |
| `omarchy-noctalia-plugin-scaffold` | building an Omarchy/Noctalia widget or Hyprland desktop tool |

`./check-skill` validates all of them; `install-skill --tui` installs interactively.

## Adding a new skill

Make a directory, write a `SKILL.md` with `name` and `description` in the
frontmatter, and put the skill body underneath. Anything else you drop in
alongside it — `references/`, `templates/`, scripts, whatever — comes along for
free, since the whole directory is what gets installed.

```bash
mkdir another-skill
$EDITOR another-skill/SKILL.md
```

## Validating skills

```bash
./check-skill              # all skills
./check-skill modern-tui   # just one
./check-skill --strict     # treat warnings as failures too
```

It checks the boring stuff so I don't have to: frontmatter actually parses as YAML
and has `name`/`description`, `name` matches the directory it's in, every
`references/…` path mentioned in `SKILL.md` really exists (and vice versa — no
orphaned reference files), markdown fences are balanced, and the description
doesn't blow Hermes' 57-character skill-index preview on boilerplate.

It also warns about the things that quietly rot: a missing
`version`/`license`/`platforms`, a missing `metadata.hermes.category` (without it
`--target hermes` dumps the skill in `general/`), and any `~/Projects/…` path a skill
mentions that doesn't exist on this machine — that last one catches "grep the sibling
repo" instructions pointing at a repo that was never cloned here. Mentions already
written defensively (guarded with `2>/dev/null`, or globbed) are left alone.

Exits non-zero on errors; warnings won't fail the run unless you pass `--strict`.

## Installing skills

Run `install-skill` from wherever — it figures out its own location, so your
current directory doesn't matter.

```bash
# Install one skill everywhere (agents, codex, claude)
~/Projects/skillz/install-skill modern-tui

# Install a few at once
~/Projects/skillz/install-skill modern-tui another-skill

# Just one target
~/Projects/skillz/install-skill modern-tui --target codex

# Same as the default, just spelled out
~/Projects/skillz/install-skill modern-tui --all

# See what's here and where it's already installed
~/Projects/skillz/install-skill --list

# Help
~/Projects/skillz/install-skill --help
```

If you put `~/Projects/skillz` on your `PATH` (or symlink `install-skill` onto
it), you can drop the leading path and just type `install-skill modern-tui`.

## Or skip the flags and just browse

```bash
install-skill --tui
# or, at an interactive prompt, just:
install-skill
```

Typing `install-skill` with no arguments launches the picker automatically
when you're at a real prompt (both stdin and stdout are a TTY). Piped or
scripted invocations with no args still get the old behavior — usage text
and a non-zero exit — so nothing that depends on that fails silently.

This is the same installer, just interactive — a little Bubble Tea app (built
with `modern-tui`'s own house style, naturally) that lists every skill in the
repo, shows install status per target at a glance, and lets you toggle a
skill on or off for `agents`/`codex`/`claude`/`hermes` with a couple of
keystrokes instead of remembering flags:

- `↑`/`↓` or `j`/`k` to move, `enter` to open a skill
- inside a skill, `space` installs at the highlighted target, or asks
  `y`/`n` to confirm before uninstalling one that's already there
- `r` re-scans status, `esc`/`q` backs out, `q` quits from the list

It's read-only about anything risky: a target it doesn't recognize as either
"not installed" or "installed by this repo" is reported as a conflict rather
than touched, same as the flag-driven installer without `--force`.

The picker scales to the full list without any changes as more skills land.

The first run builds `tui/skillz-tui` automatically (needs Go on `PATH`);
after that it just launches. If you'd rather build it yourself:

```bash
cd ~/Projects/skillz/tui
go build -o skillz-tui .
```

## How installs actually work

`~/Projects/skillz` is the source of truth. For any target that just scans a
directory for `*/SKILL.md`, the installer symlinks straight back into this
repo — so editing a skill here shows up everywhere it's installed immediately,
no reinstall needed.

| Target    | Destination                              | Method  | Notes |
|-----------|-------------------------------------------|---------|-------|
| `agents`  | `~/.agents/skills/<skill>`                | symlink | shared/generic skills directory a few tools already read from |
| `codex`   | `~/.codex/skills/<skill>` (`$CODEX_HOME`) | symlink | Codex's own installer does a plain directory scan, which follows symlinks fine |
| `claude`  | `~/.claude/skills/<skill>`                | symlink | Claude Code's personal skills directory |
| `hermes`  | `~/.hermes/skills/<category>/<skill>`     | copy    | opt-in only, see below |

The installer never deletes anything on its own. If something's already sitting
at the destination and isn't already a correct symlink into this repo, it
leaves it alone and tells you to pass `--force` — which moves the old thing
aside to `<dest>.bak.<timestamp>` rather than clobbering it.

Running `install-skill` again on something already installed is a no-op; it
notices the correct symlink is already there and just says so.

### Hermes is the one exception — opt-in, and it copies instead of linking

Hermes manages its own `~/.hermes/skills/` tree with a curator process
(`.curator_ledger.jsonl`, `.bundled_manifest`, `.usage.json`) and files things
into categories like `software-development/`. I didn't want the installer
messing with that automatically, so:

- it only touches Hermes when you pass `--target hermes` explicitly — it's
  left out of `--all` and out of the no-flags default,
- it copies the skill directory into `~/.hermes/skills/<category>/<skill>`
  rather than symlinking, sorting by the `metadata.hermes.category` field in
  the skill's frontmatter (or `general` if that's missing),
- it never touches `~/.hermes/config.yaml`. If you want Hermes reading
  `~/.agents/skills` directly instead, that's a config change you make
  yourself.

Because it's a copy, editing a skill here won't update the Hermes side by
itself — rerun `install-skill <skill> --target hermes` after making changes,
or check in with Hermes' own skill list/refresh.

## Updating the repo

```bash
cd ~/Projects/skillz
git pull
```

Since installs are symlinks, pulling changes to an existing skill is live
everywhere instantly — nothing to reinstall. If the pull brings in a brand
new skill directory, that one still needs a one-time
`install-skill <new-skill>` to get its symlinks created.

## Supported agents / targets

- **Codex** — works: `$CODEX_HOME/skills/<name>/SKILL.md` is a flat scan that
  follows symlinks without complaint.
- **Claude Code** — works: personal skills load from
  `~/.claude/skills/<name>/SKILL.md`, and it picked up a symlinked skill the
  moment it was installed.
- **`~/.agents/skills`** — a shared directory a few tools on this machine
  already treat as an external skill source (Hermes' own config even mentions
  it as an example). Installed the same way as Codex/Claude.
- **Hermes** — works, but it's the opt-in/copy-based path described above.
