# Gotchas

## Cache busting
- Add `?v=<timestamp>` to every `<script>`/`<link>` in `index.html`; bump it on each deploy
  (`sed -i -E "s/\?v=[0-9]+/?v=$V/g" index.html`).
- In the root `.htaccess` set `Cache-Control: no-cache, must-revalidate` for html/js/css/svg
  and turn `ExpiresActive Off`. Do this on the **first** deploy: phones that cached a
  7-day copy stay stale until it expires.
- Service worker: `updateViaCache: 'none'`, `fetch(req, {cache: 'no-cache'})`,
  `skipWaiting` + `clients.claim`, delete old caches on activate, and reload once on
  `controllerchange`. Use a new `CACHE` name per deploy.
- Never `pkill -f "<string that is also in your own command line>"` from a tool shell:
  it kills the shell itself (exit 144). Use `pkill -f "^node server.js"`.

## Keyboard over bottom sheets
- `<meta name="viewport" content="..., interactive-widget=resizes-content">`.
- Put the field that needs the keyboard (and its scan button) at the TOP of the form.

## Barcode scanning
- Android Chrome: `BarcodeDetector` with formats `ean_13, ean_8, upc_a, upc_e` only.
  Including `code_128` makes it read lot/plant codes on cans.
- iOS (all browsers are WebKit): no `BarcodeDetector`. Bundle ZXing
  (`@zxing/library` UMD, ~330 KB) locally. Do NOT use its `decodeFromConstraints` loop
  or `reader.decode(video)` when the video is styled (`object-fit: cover`,
  `width: 100%`) — they miss codes. Own your loop: draw `video` to a canvas at
  `videoWidth × videoHeight`, then
  `reader.decodeBitmap(new BinaryBitmap(new HybridBinarizer(new HTMLCanvasElementLuminanceSource(canvas))))`
  every ~120 ms.
- Accept a code only if it is 8/12/13 digits, the UPC/EAN check digit is valid, and the
  same code is seen on 2 reads in a row.
- Camera needs https (or localhost). Always keep a typed-number fallback.
- Test without a camera: ffmpeg + Pillow can render an EAN-13 PNG into a `.y4m`; run
  Chromium with `--use-fake-device-for-media-stream --use-fake-ui-for-media-stream
  --use-file-for-fake-video-capture=bar.y4m`.

## Product lookup
- Open Food Facts is volunteer-run and had a 502 outage; try `.org`, then `.net`, then
  UPCitemdb trial (`/prod/trial/lookup?upc=`, ~100/day, no browser CORS).
- Do it in the PHP/Node server, require a code length of 6–14 digits, tidy ALL-CAPS
  titles, and drop brands like `n/a`. Remember scanned codes in your own data.
- Some products are in no free database; the typed name is remembered per barcode.

## Debugging "it looks different on my phone"
- Compare `curl` of the live HTML with local; then screenshot with a fresh headless
  browser. If both are right, it is the user's cached copy — check the version stamp.
- Drive Chromium over CDP from plain Node (Node ≥ 22 has `WebSocket`) for scripted
  screenshots at 360/412/768/1440 px, light and dark.
