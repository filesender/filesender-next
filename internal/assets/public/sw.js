/// <reference lib="webworker" />

// These are to make sure the service worker directly works on first time connection
self.addEventListener("install", () => {
    self.skipWaiting();
});
self.addEventListener("activate", event => {
    event.waitUntil(clients.claim());
});

const downloads = new Map();

// When receiving a message from the client
self.addEventListener("message", e => {

    // When receiving new file info
    if (e.data.type === "download") {
        const broadcast = new BroadcastChannel(e.data.id);
        downloads.set(e.data.id, {
            broadcast,
            fileName: e.data.fileName,
            stream: e.data.stream
        });

        broadcast.postMessage({
            type: "downloadAvailable",
            id: e.data.id
        });
    }
});

self.addEventListener("fetch", e => {
    const url = new URL(e.request.url);
    console.log('Intercepting fetch:', url.pathname);

    if (url.pathname.includes('/download/') && !url.pathname.includes('/api/v')) {
        const id = url.pathname.split('/').pop();
        const download = downloads.get(id);

        if (download) {
            e.respondWith(new Response(download.stream, {
                headers: {
                    'Content-Type': 'application/octet-stream',
                    'Content-Disposition': `attachment; filename="${download.fileName}"`
                }
            }));
            download.delete(id);
        }
    }
});
