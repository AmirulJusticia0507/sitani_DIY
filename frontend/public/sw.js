// Service worker SiTani DIY.
//
// Prinsip: hanya aset publik yang boleh disimpan. Navigasi, balasan API
// (/v1/*), rute autentikasi SSO (/auth/*), dan dokumen pengguna tidak pernah
// di-cache agar data pribadi atau sesi tidak tersimpan usang di perangkat.
const CACHE_NAME = "sitani-static-v2";
const OFFLINE_URL = "/offline";
const STATIC_ASSETS = [
  "/favicon.svg",
  "/apple-touch-icon.png",
  "/icons/icon-192.png",
  "/icons/icon-512.png",
  "/icons/icon-maskable-512.png",
  OFFLINE_URL,
];

self.addEventListener("install", (event) => {
  event.waitUntil(
    caches
      .open(CACHE_NAME)
      // cache: "reload" memaksa aset precache diambil dari jaringan, bukan dari
      // cache HTTP yang mungkin sudah lama.
      .then((cache) => cache.addAll(STATIC_ASSETS.map((url) => new Request(url, { cache: "reload" })))),
  );
  self.skipWaiting();
});

self.addEventListener("activate", (event) => {
  event.waitUntil(
    caches.keys().then((keys) => Promise.all(
      keys.filter((key) => key.startsWith("sitani-static-") && key !== CACHE_NAME)
        .map((key) => caches.delete(key)),
    )),
  );
  self.clients.claim();
});

self.addEventListener("fetch", (event) => {
  const request = event.request;
  const url = new URL(request.url);

  if (request.method !== "GET" || url.origin !== self.location.origin) return;

  // Navigasi: jaringan lebih dulu agar pembaruan langsung terlihat. Bila
  // jaringan gagal (offline), tampilkan halaman /offline yang sudah di-precache
  // alih-alih halaman galat peramban. HTML halaman lain tidak disimpan supaya
  // isi yang bergantung sesi tidak tersaji usang.
  if (request.mode === "navigate") {
    event.respondWith(
      fetch(request).catch(() =>
        caches.match(OFFLINE_URL).then((cached) => cached || Response.error()),
      ),
    );
    return;
  }

  // Sisa penanganan hanya untuk aset statis publik yang eksplisit.
  if (!url.pathname.startsWith("/_next/static/") && !STATIC_ASSETS.includes(url.pathname)) return;

  event.respondWith(
    caches.match(request).then((cached) => cached || fetch(request).then((response) => {
      if (response.ok && response.type === "basic") {
        const copy = response.clone();
        void caches.open(CACHE_NAME).then((cache) => cache.put(request, copy));
      }
      return response;
    })),
  );
});
