// Penlight Quiz Service Worker - Runtime Image Caching (Ref: ADR-0007, ADR-0008)
const IMAGE_CACHE_NAME = 'penlight-images-v1';

self.addEventListener('install', () => {
  self.skipWaiting();
});

self.addEventListener('activate', (event) => {
  event.waitUntil(self.clients.claim());
});

self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url);

  // Intercept immutable image requests (/images/*)
  if (url.pathname.startsWith('/images/')) {
    event.respondWith(
      caches.open(IMAGE_CACHE_NAME).then(async (cache) => {
        // 1. Cache-First: return if already cached
        const cachedResponse = await cache.match(event.request);
        if (cachedResponse) {
          return cachedResponse;
        }

        // 2. Network Fetch & Runtime Cache (past displayed images retention)
        try {
          const networkResponse = await fetch(event.request);
          if (networkResponse && networkResponse.status === 200) {
            cache.put(event.request, networkResponse.clone());
          }
          return networkResponse;
        } catch (_fetchErr) {
          return (
            cachedResponse ||
            new Response(null, {
              status: 504,
              statusText: 'Gateway Timeout (Offline)',
            })
          );
        }
      }),
    );
  }
});
