<?php
// Mary's Pantry — shared data endpoint for shared hosting (replaces server.js).
// Needs api/config.php (created by deploy.sh) containing: <?php return ['key_sha256' => '...'];
header('Cache-Control: no-store');
header('X-Content-Type-Options: nosniff');

$cfg = is_file(__DIR__ . '/config.php') ? require __DIR__ . '/config.php' : [];
$hash = $cfg['key_sha256'] ?? '';
$given = $_SERVER['HTTP_X_PANTRY_KEY'] ?? '';
if ($hash === '' || !hash_equals($hash, hash('sha256', $given))) {
    http_response_code(401);
    exit('locked');
}

// Keep data outside the web root when possible (…/www/mary -> …/pantry-data).
$dir = $cfg['data_dir'] ?? dirname(__DIR__, 2) . '/pantry-data';
if (!is_dir($dir) && !@mkdir($dir . '/backups', 0700, true)) {
    $dir = __DIR__ . '/store'; // fallback, protected by .htaccess
    @mkdir($dir . '/backups', 0700, true);
}
@mkdir($dir . '/backups', 0700, true);
$store = $dir . '/data.json';

if ($_SERVER['REQUEST_METHOD'] === 'GET') {
    if (!is_file($store)) { http_response_code(404); exit; }
    header('Content-Type: application/json');
    readfile($store);
    exit;
}

if ($_SERVER['REQUEST_METHOD'] === 'PUT') {
    $body = file_get_contents('php://input');
    if (strlen($body) > 20000000 || json_decode($body) === null) { http_response_code(400); exit('bad json'); }
    if (is_file($store)) {
        $day = $dir . '/backups/data-' . date('Y-m-d') . '.json';
        if (!is_file($day)) {
            copy($store, $day);
            $files = glob($dir . '/backups/data-*.json');
            sort($files);
            while (count($files) > 30) unlink(array_shift($files));
        }
    }
    $tmp = $store . '.tmp';
    file_put_contents($tmp, $body, LOCK_EX);
    rename($tmp, $store);
    http_response_code(204);
    exit;
}

http_response_code(405);
