// Barcode scanning (native BarcodeDetector, bundled ZXing fallback for iPhone, manual entry) + product lookup.
(function () {
  const { esc } = L;

  // Only trust a retail barcode (UPC/EAN with a valid check digit).
  const validRetail = c => {
    if (!/^\d+$/.test(c) || ![8, 12, 13].includes(c.length)) return false;
    const d = c.split('').map(Number), chk = d.pop();
    const sum = d.reverse().reduce((a, n, i) => a + n * (i % 2 === 0 ? 3 : 1), 0);
    return (10 - (sum % 10)) % 10 === chk;
  };

  let zxingLoading;
  function loadZXing() {
    if (window.ZXing) return Promise.resolve();
    return zxingLoading || (zxingLoading = new Promise((res, rej) => {
      const el = document.createElement('script'); el.src = 'js/vendor/zxing.min.js?v=1790952592';
      el.onload = res; el.onerror = () => { zxingLoading = null; rej(new Error('load')); }; document.head.appendChild(el);
    }));
  }

  // Starts the camera in `video` and calls onGood(code) once a valid code is seen on 2 reads in a row.
  // Uses the browser's BarcodeDetector where it exists (Android Chrome), else the bundled ZXing reader (iPhone/Safari).
  async function startCamera(video, onGood) {
    let last = '', hits = 0, stopped = false;
    const seen = code => { if (!validRetail(code)) return; hits = code === last ? hits + 1 : 1; last = code; if (hits >= 2) { stopped = true; onGood(code); } };

    if ('BarcodeDetector' in window) {
      const det = new BarcodeDetector({ formats: ['ean_13', 'ean_8', 'upc_a', 'upc_e'] });
      const stream = await navigator.mediaDevices.getUserMedia({ video: { facingMode: 'environment' } });
      video.srcObject = stream; await video.play();
      let raf;
      const tick = async () => {
        if (stopped) return;
        try { (await det.detect(video)).forEach(x => seen(x.rawValue)); } catch {}
        raf = requestAnimationFrame(tick);
      };
      tick();
      return () => { stopped = true; cancelAnimationFrame(raf); stream.getTracks().forEach(t => t.stop()); };
    }

    await loadZXing();
    const Z = window.ZXing, hints = new Map();
    hints.set(Z.DecodeHintType.POSSIBLE_FORMATS, [Z.BarcodeFormat.EAN_13, Z.BarcodeFormat.EAN_8, Z.BarcodeFormat.UPC_A, Z.BarcodeFormat.UPC_E]);
    hints.set(Z.DecodeHintType.TRY_HARDER, true);
    const reader = new Z.BrowserMultiFormatReader(hints);
    video.setAttribute('playsinline', ''); video.muted = true;
    const stream = await navigator.mediaDevices.getUserMedia({ video: { facingMode: { ideal: 'environment' }, width: { ideal: 1280 }, height: { ideal: 720 } } });
    video.srcObject = stream; await video.play();
    let timer;
    const canvas = document.createElement('canvas'), ctx = canvas.getContext('2d', { willReadFrequently: true });
    const tick = () => {
      if (stopped) return;
      try {
        if (video.videoWidth) {
          // Copy the frame at its native size so how the video is displayed doesn't matter.
          canvas.width = video.videoWidth; canvas.height = video.videoHeight;
          ctx.drawImage(video, 0, 0);
          const bitmap = new Z.BinaryBitmap(new Z.HybridBinarizer(new Z.HTMLCanvasElementLuminanceSource(canvas)));
          seen(reader.decodeBitmap(bitmap).getText());
        }
      } catch {}   // throws when there's no code in this frame
      timer = setTimeout(tick, 120);
    };
    tick();
    return () => { stopped = true; clearTimeout(timer); stream.getTracks().forEach(t => t.stop()); try { reader.reset(); } catch {} };
  }

  function open(onCode) {
    const canCamera = !!(navigator.mediaDevices && navigator.mediaDevices.getUserMedia);
    let stopCam = null, stopped = false;
    const s = U.openSheet(`<div class="row between"><h3 class="sheet-title" style="margin:0">Scan a barcode</h3><button class="iconbtn" data-close aria-label="Close">${U.I.x}</button></div>
      <div class="row gap" style="margin:12px 0"><input type="text" id="manual" inputmode="numeric" placeholder="Or type the barcode number"><button class="btn primary sm" id="go">Use</button></div>
      ${canCamera ? `<div class="scanbox"><video id="vid" playsinline muted autoplay></video></div><p class="muted small-t" id="hint" style="margin:10px 0 0">Point the camera at the main barcode — the long one with numbers underneath. Hold steady.</p>` : `<p class="muted">${window.isSecureContext ? "This browser can't use the camera" : 'The camera needs a secure (https) connection'} — type the number above instead.</p>`}`,
      { className: 'top' });
    const stop = () => { stopped = true; if (stopCam) stopCam(); };
    const done = code => { stop(); s.close(); setTimeout(() => onCode(code), 240); };
    s.el.addEventListener('click', e => { if (e.target.closest('[data-close]') || e.target.classList.contains('scrim')) stop(); });
    s.body.querySelector('#go').onclick = () => { const v = s.body.querySelector('#manual').value.trim(); if (v) done(v); };
    s.body.querySelector('#manual').addEventListener('keydown', e => { if (e.key === 'Enter') s.body.querySelector('#go').click(); });
    if (!canCamera) return;
    startCamera(s.body.querySelector('#vid'), code => { U.haptic(20); done(code); })
      .then(fn => { if (stopped) fn(); else stopCam = fn; })
      .catch(() => { const h = s.body.querySelector('#hint'); if (h) h.textContent = "Couldn't start the camera. Allow camera access for this site, or type the number above."; });
  }

  async function lookup(code) {
    // Our own server does the lookup (tries several databases, no browser restrictions).
    try {
      const r = await fetch(S.base() + 'api/lookup.php?code=' + encodeURIComponent(code), { headers: S.hdr() });
      if (r.ok) { const j = await r.json(); if (j && j.name) return j; }
    } catch {}
    try {
      const r = await fetch(`https://world.openfoodfacts.org/api/v2/product/${encodeURIComponent(code)}.json?fields=product_name,brands`);
      const j = await r.json();
      if (j.status === 1 && j.product.product_name) return { name: j.product.product_name, brand: (j.product.brands || '').split(',')[0].trim() };
    } catch {}
    return null;
  }

  window.Scan = { open, lookup };
})();
