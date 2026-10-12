<?php
// Barcode -> product name. Tries Open Food Facts (two mirrors), then UPCitemdb.
header('Content-Type: application/json');
header('Cache-Control: private, max-age=86400');
$cfg = is_file(__DIR__ . '/config.php') ? require __DIR__ . '/config.php' : [];
$hash = $cfg['key_sha256'] ?? '';
if ($hash === '' || !hash_equals($hash, hash('sha256', $_SERVER['HTTP_X_PANTRY_KEY'] ?? ''))) { http_response_code(401); exit('{}'); }

$code = preg_replace('/\D/', '', $_GET['code'] ?? '');
if (strlen($code) < 6 || strlen($code) > 14) { echo '{}'; exit; }

function get($url) {
    $ua = 'MarysPantry/1.0';
    if (function_exists('curl_init')) {
        $c = curl_init($url);
        curl_setopt_array($c, [CURLOPT_RETURNTRANSFER => 1, CURLOPT_TIMEOUT => 6, CURLOPT_FOLLOWLOCATION => 1, CURLOPT_USERAGENT => $ua]);
        $r = curl_exec($c); $ok = curl_getinfo($c, CURLINFO_HTTP_CODE) === 200; curl_close($c);
        return $ok ? $r : null;
    }
    $r = @file_get_contents($url, false, stream_context_create(['http' => ['timeout' => 6, 'header' => "User-Agent: $ua\r\n"]]));
    return $r === false ? null : $r;
}
function tidy($s) {
    $s = trim($s);
    if ($s !== '' && $s === strtoupper($s)) $s = ucwords(strtolower($s));
    return $s;
}

foreach (['https://world.openfoodfacts.org', 'https://world.openfoodfacts.net'] as $base) {
    $r = get("$base/api/v2/product/$code.json?fields=product_name,generic_name,brands,quantity");
    $j = $r ? json_decode($r, true) : null;
    if ($j && ($j['status'] ?? 0) == 1) {
        $p = $j['product'];
        $name = $p['product_name'] ?? '' ?: ($p['generic_name'] ?? '');
        if ($name !== '') { echo json_encode(['name' => tidy($name), 'brand' => ($b = tidy(explode(',', $p['brands'] ?? '')[0])) && strtolower($b) !== 'n/a' ? $b : '', 'size' => $p['quantity'] ?? '']); exit; }
    }
}
$r = get('https://api.upcitemdb.com/prod/trial/lookup?upc=' . $code);
$j = $r ? json_decode($r, true) : null;
if ($j && !empty($j['items'][0]['title'])) {
    $i = $j['items'][0];
    echo json_encode(['name' => tidy($i['title']), 'brand' => tidy($i['brand'] ?? ''), 'size' => $i['size'] ?? '']); exit;
}
echo '{}';
