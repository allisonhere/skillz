---
name: go-cli-workflow
description: Use when building, testing, releasing, or scaffolding one of the user's Go CLI repos (assho, z13ctl, z13control, murmur, omarkey, tidemail) — covers the preflight command order, tag-based versioning gotchas, and multi-module traps.
version: 1.0.0
license: MIT
platforms: [linux, macos]
metadata:
  hermes:
    tags: [go, cli, makefile, goreleaser, versioning, release]
    category: software-development
---

# Go CLI workflow

These repos follow a consistent shape. Before treating a change as done, run through
this checklist rather than guessing at project-specific commands.

## Pre-push checklist

Run these in order (mirrors what CI runs, from tidemail/CONTRIBUTING.md):

```sh
gofmt -w .
go build ./...
go vet ./...
go test ./... -race
golangci-lint run
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
```

`govulncheck` is not part of CI, so nothing else will catch it. It separates
vulnerabilities your code actually *calls* ("Your code is affected by N") from ones that
only exist in required modules; fix the first group before a release and treat the second
as a low-urgency bump. In `tidemail` it found 8 reachable issues in `x/net` (reached via
`goquery`'s HTML parsing) and `goldmark`; bumping with `go get mod@version && go mod tidy`
cleared them. `go vet` and the tests missed all of it.

If a `Makefile` exists, prefer its targets over raw commands — they sometimes wrap extra
setup. **But don't assume one exists.** Verified across the local clones: `assho`,
`z13ctl`, `z13control`, and `whatthedock` have a Makefile; `tidemail`, `tideui`,
`tideftp`, `pinforge`, `tide`, `ripple`, `murmur`, and `omarkey` do not. And where one
exists the targets may not be what you expect — `assho/Makefile` exposes
`all build install uninstall run clean` with **no** `lint`, `test`, or `cover` target.
Check first:

```sh
grep -E '^[a-z][a-z-]*:' Makefile | cut -d: -f1
```

`scripts/preflight.sh` in this skill runs the whole checklist above, including the extra
module loop described below, and lists the Makefile targets it found.

### Getting `golangci-lint` to zero

`tide` went from 94 issues to 0 in three commits. What made it tractable:

- `golangci-lint run --new-from-rev=<last-tag> ./...` separates what you introduced (24 of
  the 94) from what was already there. A failing lint step stops the `deploy.sh` pipeline, so
  inherited issues block a release too.
- The linter caps how many repeats of one finding it shows, so fixing some reveals more
  (94 → 65 → 44 → 17 → 5 → 19). Set `issues: {max-issues-per-linter: 0, max-same-issues: 0}`
  in `.golangci.yml` and loop until `golangci-lint run` **exits 0**, not until the visible list
  is empty.
- Pin a config (`version: "2"`) that excludes errcheck only for calls whose error is
  conventionally ignored: `fmt.Fprint*`, `(io.ReadCloser).Close`, `(*database/sql.Rows).Close`,
  `(*database/sql.DB).Close`, `(*database/sql.Tx).Rollback`, plus errcheck in `_test.go`. Discard
  the rest explicitly (`_ = f.Close()`). **Handle** the close of a file you wrote: a bare
  `defer f.Close()` after an encoder hid a failed `config.Save`. Errors dropped on a save path
  should reach the user ("could not save settings"), not vanish.
- `unused` counts test usage, so anything it reports is dead everywhere: delete it. A small
  `go/parser` program that removes each `FuncDecl` plus its doc comment beats hand-editing 27
  functions; run the tests afterwards.
- `SA1019` on `lipgloss.Style.Copy()`: styles are value types, so just drop the call.
- zsh does not word-split an unquoted `$VAR` into arguments, so `tool $ARGS` passes one
  argument; pipe the list through `xargs`.

## Fast local loop

For small edits, use a targeted loop while iterating, then run the full preflight before
calling the change done:

```sh
gofmt -w ./path/to/changed.go
go test ./internal/package -run TestName
go test ./internal/package -race
```

Before deciding the package path, confirm module/workspace context:

```sh
go env GOMOD
go env GOWORK
find . -maxdepth 3 -name go.mod -print
```

If `GOWORK` is set or nested `go.mod` files exist, be explicit about which module you're
testing. A green root `go test ./...` does not prove a sibling/nested module was tested.

**Pitfall — bogus `typecheck` errors from golangci-lint.** If `golangci-lint run` fails
with stdlib errors you didn't cause, e.g.

```text
/usr/lib/go/src/crypto/internal/randutil/randutil.go:11:2: could not import math/rand/v2
  (method must have no type parameters) (typecheck)
```

the linter binary was built against a different Go than the installed toolchain
(`golangci-lint --version` vs `go version`). Reinstall golangci-lint for the current Go
before believing the failure — `go build`/`go vet`/`go test` passing while only
`typecheck` complains about files under `/usr/lib/go` is the tell.

## Versioning

Version comes from git tags via:

```sh
git describe --tags --match 'v*' --always --dirty
```

The `--match 'v*'` filter is load-bearing where a repo carries a second tag series.
`z13ctl` is the case that motivated it: it has an `api/` submodule (`api/go.mod`) and a
`.goreleaser.yml`. Without the filter, `git describe` (and goreleaser, if unconfigured)
picks the newest tag regardless of series and mis-versions the main binary — reporting
something like `api/v1.1.7` as its own version.

Confirm the situation in the repo you're in rather than assuming:

```sh
git tag --list | sed 's/[0-9].*//' | sort -u     # which tag series exist
ls api/go.mod 2>/dev/null && echo "separate module present"
grep -n 'ignore_tags' .goreleaser.yml 2>/dev/null
```

Where an `api/*` series exists alongside `v*`, `.goreleaser.yml` needs
`git.ignore_tags: ["api/*"]` (or equivalent) or the release mis-versions the main binary.

Note `z13ctl` is **local-only** — it is not on `github.com/allisonhere`, so `gh` lookups
for it will 404.

Version is injected via ldflags, typically:

```
-X <module>/cmd.Version=$(VERSION)
```

or `-X main.version=$(VERSION)` for single-package repos without a `cmd/` layout.

## Multi-module repos

If the repo has an `api/` (or similar) directory with its own `go.mod`, it's a separate
module. `go test ./...` and `go vet ./...` from the repo root will silently skip it —
run the same commands again inside that directory, or use the Makefile's `test`/`lint`
targets if they already loop over both modules (check first; some do, some don't).

## Releases

Per-repo, verified — do not generalize:

| Repo | Release mechanism |
|------|-------------------|
| z13ctl | `.goreleaser.yml` (the only local repo with one) + Makefile |
| tide | `release.sh` + `install.sh` at root, and `deploy.sh` + `deploy.conf` |
| tidemail, tidedeck, tidesms | `deploy.sh` + `deploy.conf` + `.github/workflows/release.yml` (see `tide-tui-ecosystem`) |
| murmur | `install.sh` only |
| ripple, omarkey | neither — `ripple` is an importable component, not a shipped binary |
| assho, z13control, whatthedock | Makefile (`make install` → `/usr/local/bin`, man pages under `/usr/local/share/man/man1`, via `sudo`, never on build) |

Check before reaching for goreleaser:

```sh
ls .goreleaser.y*ml release.sh install.sh Makefile 2>/dev/null
```

## Before tagging a release

A pass that caught real problems in `tidemail` v1.4.0 — run it in this order:

1. Be on `main` (or merge to it first). Check `git status` for strays; `deploy.sh` commits
   everything with `git add -A`.
2. Full preflight (`scripts/preflight.sh`): gofmt, build, vet, `go test -race`,
   `golangci-lint`, `govulncheck`. CI's `golangci-lint` is `latest`, so it flags things like
   `errcheck` on `fmt.Sscanf`, dead initializers (`ineffassign`), and `math.Pow(x, 3)`
   (`QF1005`) that vet and tests never mention.
3. Docs consistency: anything user-facing you changed must also be in the README, the
   guide, in-app help, and `STATUS.md` (check the repo's `AGENTS.md`/`CONTRIBUTING.md`).
4. `CHANGELOG.md`: entries under `## Unreleased`, or a populated `## <tag>` section — the
   release workflow publishes the section named after the tag, verbatim.
5. Dry-run the notes extraction the workflow uses
   (`awk -v ver="## vX.Y.Z" '$0==ver{f=1;next} /^## /&&f{f=0} f' CHANGELOG.md`).
6. After the tag: confirm `main` contains it (`git branch -r --contains vX.Y.Z`).

Rehearse the console first. `./deploy.sh --check` prints the readiness checklist with no
terminal; driving `./deploy.sh --dry-run` → Full release in tmux walks check, bump, test, lint,
build, commit, push and tag without committing anything. It **still rebuilds the tracked
`dist/` binaries**, so run `git checkout -- dist` afterwards. `RELEASE_NOTES="generated"` in
`deploy.conf` (`tide`) skips the notes step, so rename `## Unreleased` to the new version by
hand; with `changelog` (`tidemail`) the console promotes it.

## A headless command beside a TUI

`tide --refresh`, `--unread-count` and `--bar` run without the reader, from cron or a status bar.
What it took to do that safely:

- Dispatch headless flags in `main` **before** config, database and terminal setup, so nothing
  writes escape codes to stdout. A flag the binary does not know falls through to the TUI, which
  prints terminal sequences instead of an error.
- Put the work both paths need (fetch, convert, store) in a package neither UI owns
  (`internal/refresh`) and have the TUI call it; the feed manager's add path had grown its own copy.
- The database must tolerate a second process: WAL plus a busy timeout. With `modernc.org/sqlite`
  put it in the DSN (`path + "?_pragma=busy_timeout(5000)"`), because a `PRAGMA` through `db.Exec`
  reaches only one pooled connection. Test it: hold a write transaction on one handle, commit it
  from a timer, and write from a second handle; without the timeout that write fails at once with
  SQLITE_BUSY.
- When another program runs your binary (a bar, cron), **probe `exe --flag` for the expected
  output before wiring it in.** An older build ignores the flag, starts the TUI, and its startup
  escape codes (`OSC 11`, Kitty `a=d`) became the bar's text. A toggle that falls back to the
  installed binary when run from `go run` hits this exactly when that binary is old.
- `//go:embed CHANGELOG.md` plus a test that the embedded text has the sections you expect keeps
  an in-app changelog in step with the build.

## Architecture docs

`z13ctl` is the repo with a root `CLAUDE.md` (and a `CONTRIBUTING.md`); `tidemail` keeps
`CONTRIBUTING.md` + `STATUS.md`; `tide` keeps `STATUS.md`, `PLAN.md`, `folders.md`, and a
`.codex` directory. Most of the others carry only a README.

The rule still holds: when a change is structural — new package, moved responsibility,
new subcommand — update whichever of those docs the repo actually has, in the same
change, not as an afterthought. It's the thing future sessions (including yours) read
first. Don't create a `CLAUDE.md` in a repo that never had one — match the repo's
existing convention.

## Conventions worth preserving

- **Error handling**: surface failures to the user rather than dropping them silently.
  If an error is deliberately ignored, say so: `_ = f.Close()`, or a `//nolint:errcheck`
  with a short reason. Calls whose error is conventionally ignored are better excluded once
  in `.golangci.yml` than annotated everywhere (see "Getting `golangci-lint` to zero").
- **Tests**: table-driven by default.
- Man pages / install paths: tools that install system-wide (see assho's Makefile)
  install to `/usr/local/bin` and `/usr/local/share/man/man1` via `sudo`, gated behind
  an explicit `make install` — never assume install-on-build.

## See also

- `tide-tui-ecosystem` — cross-repo facts for the Tide-family Go TUIs (duplicated files,
  `tideui` as a dependency, per-repo release scripts).
- `modern-tui` — how to design and implement the TUI itself, when the CLI has one.
