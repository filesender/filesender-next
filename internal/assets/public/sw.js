/// <reference lib="webworker" />

// These are to make sure the service worker directly works on first time connection
self.addEventListener("install", () => {
    self.skipWaiting();
});
self.addEventListener("activate", event => {
    event.waitUntil(clients.claim());
});

/**
 * @typedef {Object} FileInfo
 * @property {boolean} available
 * @property {boolean} chunked
 * @property {number} chunkCount
 * @property {string} userId
 * @property {string} fileId
 */
/** @type {FileInfo} */
var fileInfo;

// When receiving a message from the client
self.addEventListener("message", e => {

    // When receiving new file info
    if (e.data.type === "info") {
        fileInfo = e.data.info;
    }
});
