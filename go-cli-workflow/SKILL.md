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

If a `Makefile` exists (it almost always does), prefer its targets over raw commands —
`make lint`, `make test`, `make cover` — since they sometimes wrap extra setup.

## Versioning

Version comes from git tags via:

```sh
git describe --tags --match 'v*' --always --dirty
```

The `--match 'v*'` is load-bearing, not decorative: repos like z13ctl also carry an
`api/vX.Y.Z` tag series for a separate submodule. Without the filter, `git describe`
(and goreleaser, if unconfigured) picks up the newest tag regardless of series and
mis-versions the main binary — it'll report `api/v1.1.7` as its own version. If you see
a `.goreleaser.yml`, check it has a matching `git.ignore_tags: ["api/*"]` (or equivalent)
whenever an `api/` submodule tag series exists alongside main `v*` tags.

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

Cut from `v*` git tags. Either:
- `goreleaser release` (check for `.goreleaser.yml`), or
- a plain `deploy.sh` / `install.sh` pair that builds and copies the binary.

Don't assume which one a given repo uses — check for `.goreleaser.yml` before reaching
for goreleaser commands.

## Architecture docs

Most of these repos keep a `CLAUDE.md` at the root documenting package layout and key
flows (see z13ctl's for the fullest example). When a change is structural — new
package, moved responsibility, new subcommand — update that doc in the same change,
not as an afterthought. It's the thing future sessions (including yours) read first.

## Conventions worth preserving

- **Error handling**: surface failures to the user rather than dropping them silently.
  If an error is deliberately ignored, mark it `//nolint:errcheck` with a short reason,
  don't just discard it quietly.
- **Tests**: table-driven by default.
- Man pages / install paths: tools that install system-wide (see assho's Makefile)
  install to `/usr/local/bin` and `/usr/local/share/man/man1` via `sudo`, gated behind
  an explicit `make install` — never assume install-on-build.
