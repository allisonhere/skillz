---
name: wow-addon-workflow
description: Use when writing, debugging, testing, or releasing a World of Warcraft addon for the modern-API clients (WoW Forever / Midnight) — covers secret values, taint pitfalls, Lua 5.1 testing, probing the live client, and tag-based CurseForge releases.
version: 1.0.0
license: MIT
platforms: [linux, macos]
metadata:
  hermes:
    tags: [wow, lua, addon, curseforge, taint, secret-values]
    category: software-development
---

# WoW addon workflow

Lessons from building Tempus (unit frames, nameplates, buffs) and SpellPower (chat spell
checker). Read the pitfalls before debugging an in-game error; most of them cost a round trip.

## The client

"WoW Forever" (the install folder may still be called `_classic_beta_`) runs the **modern
Midnight addon API**: interface `16001`, retail-style Blizzard frames (`PlayerSpellsFrame`,
Edit Mode), a built-in cooldown manager and damage meter. Do not assume classic frame names.
Blizzard's UI source is usually **not on disk**, so find structure by probing (below).

## Secret values

In combat the client hides values (enemy health, many aura and cast fields). A secret value
cannot be compared, added, concatenated or used as a table key.

- Guard reads: `issecret(v)` before arithmetic or comparison; wrap risky calls in `pcall`.
- Booleans that are secret can still drive visuals: `region:SetAlphaFromBoolean(flag, a, b)`.
- Secret strings and numbers can still be displayed: `fontString:SetText(secret)`,
  `SetFormattedText("%s: %s", a, secret)`, `statusBar:SetValue(secret)`.
- Cooldowns and auras: prefer duration objects (`C_Spell.GetSpellCooldownDuration`,
  `cooldown:SetCooldownFromDurationObject`) over start/duration numbers.
- A table's **length** (`#ids`) is often still readable even when its contents are not
  (`C_UnitAuras.GetUnitAuraInstanceIDs`) — enough for "does this enemy have a purgeable buff".
- Let the game draw what it can: native aura containers and cooldown frames keep updating
  in combat when addons cannot read the data.

## Taint

- **Never reparent, `UnregisterAllEvents`, or write fields on Edit Mode managed frames**
  (`BuffFrame`, `DebuffFrame`, `PlayerFrame`, …). It tainted Blizzard's aura code and caused
  `GetAuraDataByIndex(): Auras cannot be accessed when secret while tainted by '<addon>'`
  on level-up, because Edit Mode re-runs layout then. Hide them with `SetAlpha(0)` instead.
- `/console taintLog 2` writes `Logs/taint.log` (readable from outside the game), but it
  **changes behaviour**: with it on, `hooksecurefunc` wrappers on Blizzard methods threw
  `attempt to call a nil value` (micro buttons' `Enable`, tracker `EndLayout`). Use it for a
  short capture, then turn it off with `/run SetCVar("taintLog", 0)`. If `/console ...` just
  lands in chat, use the `/run SetCVar(...)` form.
- Errors keep their **first timestamp** and bump a count. After a fix, `/reload` and check
  the timestamp is new before concluding the fix failed. Lua files are only read at load.

## Probe, don't guess

Add a slash command (`/addon probe`) that records frame names, children, method names and
`C_*` namespaces into the SavedVariables table, then read it from
`WTF/Account/<id>/SavedVariables/<Addon>.lua` after a `/reload`. Cheaper than guessing frame
structure, and it works for frames that only exist once an option is enabled.

## Testing

WoW runs **Lua 5.1**. The default `lua`/`luac` on a modern distro is 5.4/5.5 and gives false
results (e.g. `attempt to assign to const variable` on a `for` variable, missing `unpack`).

```sh
scripts/lua51-check.sh            # syntax-check every .lua and run tests/*.lua under 5.1
```

Test pure logic (serializers, rules, list handling) by loading the file with a stub `T`/`SP`
namespace: `assert(loadfile("Core/Share.lua"))("Addon", T)`. Keep UI code thin so most of it
is testable this way. Check regenerated files (a changelog built from `CHANGELOG.md`) are
regenerated **before** tests run in the release script.

## Performance

A big table build on the loading path stalls the game. A 30k-word BK-tree took ~2.7 s; the
fix was a smaller working set (top 20k) built in slices of ~250 per timer tick, with lookups
searching whatever is built so far. Cheap membership (a hash set of 100k words, ~7 MB) is fine;
the expensive structure is the one to bound.

## Release (CurseForge packager)

- `.pkgmeta`: `package-as`, an `ignore:` list (README, screenshots, branding, `.github`, `tests`,
  `tools`) and `manual-changelog` pointing at `CHANGELOG.md`.
- CurseForge's packager builds from a **pushed tag**; a GitHub Actions workflow using
  `BigWigsMods/packager` does the same for GitHub releases. Neither bumps the version for you.
- A `tools/release.sh X.Y.Z [--dry-run] [--yes]` that checks `main`, a new tag, a first
  `## X.Y.Z` changelog entry, Lua 5.1 syntax and tests, then sets the `.toc` version, commits,
  tags and (after confirmation) pushes keeps this repeatable.
- Never push or tag without the user saying so; a pushed tag publishes.

## Git on Wine / NTFS drives

Files can show every mode bit as changed. Commit with `git -c core.fileMode=false ...`. Do not
run `git config <key> <other-key>` by accident — it silently sets the value.
