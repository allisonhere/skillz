#!/usr/bin/env bash
# Syntax-checks every .lua file and runs tests/*.lua under Lua 5.1 (the version WoW uses).
# Usage: scripts/lua51-check.sh [addon-root]   (defaults to the current directory)
set -uo pipefail
cd "${1:-.}" || exit 2
LUAC="$(command -v luac5.1 || true)"
LUA="$(command -v lua5.1 || command -v luajit || true)"
[ -n "$LUAC" ] && [ -n "$LUA" ] || { echo "install lua5.1 (luac5.1 + lua5.1) first" >&2; exit 2; }
fail=0
while IFS= read -r -d '' f; do
  "$LUAC" -p "$f" || { echo "SYNTAX: $f"; fail=1; }
done < <(find . -name '*.lua' -not -path './.git/*' -print0)
for t in tests/*.lua; do
  [ -e "$t" ] || continue
  if out="$(timeout 60 "$LUA" "$t" 2>&1)"; then echo "ok    $t"; else echo "FAIL  $t"; echo "$out" | tail -5 | sed 's/^/      /'; fail=1; fi
done
exit $fail
