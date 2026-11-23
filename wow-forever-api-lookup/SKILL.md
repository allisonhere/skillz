---
name: wow-forever-api-lookup
description: Use before writing or reviewing any WoW Forever (interface 16001) addon code that calls a game API, event, or Blizzard frame — find the real signature, flavor, secrecy and behaviour in the Forever branch of Blizzard's UI source and the community references instead of guessing from retail or classic memory.
version: 1.0.0
license: MIT
platforms: [linux, macos]
metadata:
  hermes:
    tags: [wow, forever, api, lookup, secret-values, wowlua-ls]
    category: software-development
---

# Looking things up for WoW Forever

Forever is the **retail client** (interface `16001`, a 12.x engine) with its own game rules. Retail
memory is wrong for specs and talents, classic memory is wrong for the API. Never write an API call
from memory. Look it up, then prove it in the live client. Say which step backed each claim.

## 1. Forever's own UI source (best source)

Gethe/wow-ui-source has a **`forever` branch**. A local clone is usually checked out on `live`
(retail), so always name the branch. Do not switch the checkout, read through git:

```bash
cd ~/Projects/wow/wow-ui-source          # clone of github.com/Gethe/wow-ui-source
git fetch origin forever                 # refresh; build is in version.txt
git show origin/forever:version.txt      # e.g. 1.60.1.70888
git grep -n 'Name = "GetItemStatDelta"' origin/forever -- Interface/AddOns/Blizzard_APIDocumentationGenerated
git grep -l  'LiteralName = "PLAYER_SWING"' origin/forever -- Interface
git show origin/forever:Interface/AddOns/Blizzard_APIDocumentationGenerated/ItemDocumentation.lua | sed -n 980,1010p
```

- `Blizzard_APIDocumentationGenerated/*Documentation.lua` is the API reference: `Name`, `Arguments`,
  `Returns`, `Events` (`LiteralName`, `Payload`), enums, and the secrecy keys (`SecretReturns`,
  `SecretArguments`, `SecretWhen...`).
- The rest of `Interface/AddOns/` shows how Blizzard itself calls the API on Forever (talent
  panels, `Blizzard_LegacySystem`, character stats). Copy its idioms.
- A function missing from `origin/forever` does not exist on Forever, even if retail has it.
  Compare with `origin/live` when unsure which side a name is on.

Related branches there: `live` (retail), `classic_era`, `classic`, `ptr`. Ketho/BlizzardInterfaceResources
also has a `forever` branch (name lists, enums, CVars).

## 2. wowlua-ls (typed API, flavors, secret values)

TradeSkillMaster/wowlua-ls (GPL-3.0; running it is fine, do not copy its stubs into an addon) ships
9,000+ typed stubs including Forever, 1,000+ typed events, flavor data and secret-value tracking.

- Release binaries: `https://github.com/TradeSkillMaster/wowlua-ls/releases` (check the sha256 in
  `checksums.txt`). Editor extensions bundle it too.
- `wowlua_ls dump-stubs > stubs.txt` lists every global with its resolved type, one per line:
  grep it to see signatures and whether a name exists.
- `wowlua_ls check path/to/addon --severity hint` lints; exit code 1 on findings (CI-friendly).
- `.wowluarc.json` in the addon root: `{ "flavors": ["forever"], "ignore": ["Libs/"] }`.
- Useful diagnostics for Forever: `wrong-flavor-api` (call not available on Forever), `deprecated`,
  `secret-arithmetic`, `secret-comparison`, `secret-condition`, `secret-table-key`,
  `secret-argument`, `secret-access`. Hovering an API shows "Flavors:" and a Secrecy section.
- Not yet verified in this setup: running the binary against Tempus. Check output against section 1
  before trusting a finding, and do not silence a diagnostic without reading the API doc.

## 3. Community references (read for technique; check each licence before copying)

- Thunderz96/forever-addon-kit: measured findings table, `data/forever_api.json` (6,045 functions,
  269 namespaces, beta build 69893), porting tools.
- Atraeau/WoW-Addons: generated API docs with signatures (no licence file: read only).
- turlockmike/pawn-forever (MIT): Forever tooltip patterns and per-spec stat weights, derived from
  GearQuest Forever (Henrik8210).
- Blizzard's Forever developer notes and class deep dives for rule changes (merged hit and crit,
  weapon DPS on feral forms, a new 16-point talent tier).

## 4. Prove it in the client

Docs list what exists, not what an addon may call in combat or on this build. Add or run a probe
(Tempus has `/tempus gearprobe`): `pcall` every call, print results, paste them back. Record the
result with the build number.

## Verified Forever facts (level 22 Druid, interface 16001)

- No retail specializations: `GetSpecialization`, `GetNumTalentTabs`, `GetTalentInfo`,
  `GetNumSkillLines` and global `GetItemStats`/`GetItemInfo` are missing. `C_SpecializationInfo`
  returns one class entry (Druid: id 1484, role DAMAGER); slots 2 and 3 are empty.
- Talents are one class tree read with `C_ClassTalents.GetActiveConfigID()` then `C_Traits`
  (`GetConfigInfo(cfg).treeIDs`, `GetTreeNodes`, `GetNodeInfo(cfg, node).currentRank`,
  `GetEntryInfo`, `GetDefinitionInfo(...).spellID`). `C_Traits.GetConfigInfo(0)` returns nothing.
- Role comes from `GetShapeshiftForm()`, talents and a manual choice, not a spec.
- `C_Item.GetItemStats(link)` returns `ITEM_MOD_*_SHORT` keys; `C_Item.GetItemStatDelta`,
  `GetItemUniquenessByID`, `IsEquippableItem` exist. Dual wield: `IsSpellKnown(674)` is false
  without it; `IsDualWielding` exists.
- Attack power follows the classic formulas: cat AP = level*2 + 2*Str + Agi - 20; humanoid
  druid AP = 2*Str - 20.
- In combat enemy auras throw "Auras cannot be accessed when secret": time your own casts from
  `UNIT_SPELLCAST_SUCCEEDED` with a duration learned out of combat. `C_Secrets.GetSpellAuraSecrecy`
  answers per spell.

## Rules

- Quote the file and line (or probe output) behind each API claim; mark the rest as unverified.
- If branch and live probe disagree, the live client wins; record the build.
- Wrap unknown calls in `pcall` and register events inside `pcall`: an unknown event throws.
