#!/usr/bin/env bash
# Usage: ./deploy.sh user@host [remote_dir] [ssh_port]
# Uploads the app and, on first run, sets the password that unlocks the shared data.
set -euo pipefail
HOST="${1:?usage: ./deploy.sh user@host [remote_dir] [ssh_port]}"
DEST="${2:-/home/allieher/www/mary}"
PORT="${3:-22}"
cd "$(dirname "$0")"
SSH="ssh -p $PORT"

rsync -avz --delete -e "$SSH" \
  --exclude 'store/' --exclude 'node_modules/' --exclude '.git/' --exclude 'server.js' \
  --exclude 'deploy.sh' --exclude 'README.md' --exclude 'api/config.php' \
  ./ "$HOST:$DEST/"

if $SSH "$HOST" "test -f '$DEST/api/config.php'"; then
  echo "Password already set (delete $DEST/api/config.php on the server to change it)."
else
  if [ -n "${PANTRY_PASSWORD:-}" ]; then PW="$PANTRY_PASSWORD"; elif [ -t 0 ]; then read -rsp "Choose a password for the pantry: " PW; echo; else echo "No password set yet — re-run in a terminal to choose one."; exit 0; fi
  H=$(printf '%s' "$PW" | sha256sum | cut -d' ' -f1)
  $SSH "$HOST" "printf '<?php return [\"key_sha256\" => \"%s\"];\n' '$H' > '$DEST/api/config.php' && chmod 600 '$DEST/api/config.php'"
  echo "Password set."
fi
echo "Done. Open your site's /mary/ address."
