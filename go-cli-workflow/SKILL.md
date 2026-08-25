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
```

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
| tide | `release.sh` + `install.sh` at root |
| murmur | `install.sh` only |
| ripple, omarkey | neither — `ripple` is an importable component, not a shipped binary |
| assho, z13control, whatthedock | Makefile (`make install` → `/usr/local/bin`, man pages under `/usr/local/share/man/man1`, via `sudo`, never on build) |

Check before reaching for goreleaser:

```sh
ls .goreleaser.y*ml release.sh install.sh Makefile 2>/dev/null
```

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
  If an error is deliberately ignored, mark it `//nolint:errcheck` with a short reason,
  don't just discard it quietly.
- **Tests**: table-driven by default.
- Man pages / install paths: tools that install system-wide (see assho's Makefile)
  install to `/usr/local/bin` and `/usr/local/share/man/man1` via `sudo`, gated behind
  an explicit `make install` — never assume install-on-build.

## See also

- `tide-tui-ecosystem` — cross-repo facts for the Tide-family Go TUIs (duplicated files,
  `tideui` as a dependency, per-repo release scripts).
- `modern-tui` — how to design and implement the TUI itself, when the CLI has one.
