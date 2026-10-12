---
name: shared-host-pwa
description: Use when building a small installable web app (PWA) that must sync data between several phones and run on cheap shared PHP hosting (cPanel/LiteSpeed) — covers the password-protected PHP data endpoint, rsync deploy script, cache busting, and phone barcode scanning.
version: 1.0.0
license: MIT
platforms: [linux, macos]
metadata:
  hermes:
    tags: [pwa, php, shared-hosting, rsync, barcode, sync, offline]
    category: software-development
---

# Shared-host PWA with cross-device sync

Pattern for a no-build, no-dependency web app (vanilla HTML/CSS/JS) used by a
household or small team on phones, stored on ordinary shared hosting. Proven on a pantry
inventory app (items, shopping list, recipes, barcode scanning).

## Architecture

- **Frontend**: static files. All state lives in one JSON document in `localStorage`,
  pushed to and pulled from the server.
- **Backend**: one PHP file (`templates/data.php`) with GET (read JSON) and PUT (write JSON).
  Auth is a shared password sent in an `X-Pantry-Key` header; the server stores only its
  SHA-256 in `api/config.php`. Data lives outside the web root
  (`dirname(__DIR__, 2) . '/pantry-data'`) with 30 daily backups.
- **Same API in dev**: a tiny Node server answers `/api/data.php` too, so the client uses
  one URL in both places. Use **relative** URLs (`api/data.php`, `sw.js`) so the app works
  under a subfolder like `/mary/`.
- **Deploy**: `templates/deploy.sh user@host [dir] [port]` = rsync (excluding secrets and
  the dev server) plus one-time password setup over ssh. Set `PANTRY_PASSWORD` to run it
  non-interactively.

## Templates (copy, then adapt names)

| File | Purpose |
|------|---------|
| `templates/data.php` | password-protected read/write JSON endpoint with daily backups |
| `templates/lookup.php` | barcode → product name with three fallback databases |
| `templates/api.htaccess` | blocks direct access to `config.php` and `*.json` (goes in `api/.htaccess`) |
| `templates/root.htaccess` | no-cache headers, no directory listing (goes in the site root as `.htaccess`) |
| `templates/sw.js` | revalidating service worker |
| `templates/scanner.js` | native + ZXing barcode scanner sheet (expects `U`, `S`, `L` helpers; see the pantry app) |
| `templates/deploy.sh` | rsync + first-run password setup |

## Sync model (simple and safe)

Every record has `id`, `updatedAt`, and a `deleted` tombstone flag; settings carry
`settingsAt`. Merge = per-id, newer `updatedAt` wins. Sync on load, every 30 s, on tab
focus, and ~600 ms after any change. Only emit a UI update when data or status really
changed, or the view re-renders (and animates) every 30 s. Handle HTTP 401 as a
"locked" state and show a password sheet.

## Things that bit us (read `references/gotchas.md`)

1. Caching: shared hosts send 7-day `Cache-Control` by default; phones kept old CSS/JS.
2. Service worker must revalidate (`fetch(req, {cache: 'no-cache'})`) and clean old caches.
3. Barcode scanning differs on Android and iPhone.
4. Free product databases have gaps and outages — do the lookup server-side with fallbacks.
5. Keyboards cover bottom sheets.

## Checklist before telling the user it is done

- Fetch the live `index.html` and confirm it matches the local one.
- Load the live URL in a fresh headless browser and screenshot it.
- Confirm the API returns 401 without the key and 200 with it, and that `config.php`
  returns 403 directly.
- Show a visible version stamp in the UI (used for cache-busting query strings, the
  service worker cache name, and a footer) so "is my phone on the new version?" is
  answerable.
