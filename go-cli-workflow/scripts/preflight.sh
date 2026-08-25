#!/usr/bin/env bash
# Go CLI preflight: format, build, vet, test, lint — across every module in the repo.
# Reports only; never rewrites your files (uses `gofmt -l`, not `gofmt -w`).
# Usage: scripts/preflight.sh [repo-root]   (defaults to the current directory)
set -uo pipefail

root="${1:-.}"
cd "$root" || exit 2
fail=0

run() {
  printf '\n\033[1m$ %s\033[0m\n' "$*"
  "$@" || { fail=1; printf '\033[31mFAILED: %s\033[0m\n' "$*"; }
}

mapfile -t mods < <(find . -name go.mod -not -path '*/vendor/*' -printf '%h\n' | sort)
if [ "${#mods[@]}" -eq 0 ]; then
  echo "no go.mod found under $root" >&2
  exit 2
fi
echo "modules: ${mods[*]}"

for m in "${mods[@]}"; do
  printf '\n\033[36m=== module %s ===\033[0m\n' "$m"
  (
    cd "$m" || exit 1
    unformatted=$(gofmt -l .)
    if [ -n "$unformatted" ]; then
      printf '\033[31mgofmt would change:\033[0m\n%s\n' "$unformatted"
      exit 1
    fi
    run go build ./...
    run go vet ./...
    run go test ./... -race
    if command -v golangci-lint >/dev/null 2>&1; then
      run golangci-lint run
    else
      echo "golangci-lint not installed — skipping (install before pushing)"
    fi
    exit "$fail"
  ) || fail=1
done

if [ -f Makefile ]; then
  printf '\n\033[36m=== Makefile targets available ===\033[0m\n'
  grep -E '^[a-z][a-z-]*:' Makefile | cut -d: -f1 | tr '\n' ' '; echo
fi

if [ "$fail" -eq 0 ]; then
  printf '\n\033[32mpreflight OK\033[0m\n'
else
  printf '\n\033[31mpreflight FAILED\033[0m\n'
fi
exit "$fail"
