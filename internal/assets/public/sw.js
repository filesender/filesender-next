/// <reference lib="webworker" />

// These are to make sure the service worker directly works on first time connection
self.addEventListener("install", () => {
    self.skipWaiting();
});
self.addEventListener("activate", event => {
    event.waitUntil(clients.claim());
});

// When receiving a message from the client
self.addEventListener("message", e => {
    const { type, data, readableStream } = e.data;
});
