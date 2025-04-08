
self.addEventListener('install', () => {
    self.skipWaiting();
});

self.addEventListener('activate', event => {
    event.waitUntil(clients.claim());
});

const streams = new Map();

self.addEventListener('message', event => {
    const { id, name, readableStream } = event.data;
    streams.set(id, [name, readableStream]);

    event.source.postMessage({
        url: `/download/${id}`
    });
});

self.addEventListener('fetch', event => {
    const url = new URL(event.request.url);
    console.log('Intercepting fetch:', url.pathname);

    if (url.pathname.startsWith('/download/')) {
        const id = url.pathname.split('/').pop();
        const stream = streams.get(id);

        if (stream) {
            event.respondWith(new Response(stream[1], {
                headers: {
                    'Content-Type': 'application/octet-stream',
                    'Content-Disposition': `attachment; filename="${stream[0]}"`
                }
            }));
            streams.delete(id);
        }
    }
});
