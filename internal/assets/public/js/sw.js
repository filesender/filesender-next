/// <reference lib="webworker" />

// These are to make sure the service worker directly works on first time connection
self.addEventListener("install", () => {
    self.skipWaiting();
});
self.addEventListener("activate", event => {
    // eslint-disable-next-line no-undef
    event.waitUntil(clients.claim());
});

const downloads = new Map();

// When receiving a message from the client
self.addEventListener("message", e => {

    if (e.data.type === "delete") {
        downloads.delete(e.data.id);
    }
    else if (e.data.type === "download" && e.data.port) {
        const { id, fileName, port } = e.data;
        const broadcast = new BroadcastChannel(id);

        const stream = new ReadableStream({
            start(controller) {
                port.onmessage = ({ data }) => {
                    if (data.done) {
                        controller.close();
                    } else if (data.chunk) {
                        try {
                            controller.enqueue(new Uint8Array(data.chunk));
                        } catch (err) {
                            
                            broadcast.postMessage({
                                type: "downloadCancelled"
                            });
                            throw err;
                        }
                    }
                };
            }
        });

        downloads.set(id, {
            broadcast,
            fileName,
            stream
        });

        broadcast.postMessage({
            type: "downloadAvailable",
            id
        });
    }
});

self.addEventListener("fetch", e => {
    const url = new URL(e.request.url);

    if (url.pathname.includes('/dl/') && !url.pathname.includes('/api')) {
        console.log("Intercepting:", url.pathname);
        const id = url.pathname.split('/').pop();
        const download = downloads.get(id);

        if (download) {
            e.respondWith(new Response(download.stream, {
                headers: {
                    'Content-Type': 'application/octet-stream',
                    'Content-Disposition': `attachment; filename="${download.fileName}"`
                }
            }));
        }
    }
});
