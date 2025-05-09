var ENC_CHUNK_SIZE = 1024 * 1024;

/**
 * @typedef {Object} FileInfo
 * @property {boolean} available
 * @property {boolean} chunked
 * @property {number} chunkCount
 * @property {string} userId
 * @property {string} fileId
 * @property {string} fileName
 */

// eslint-disable-next-line no-unused-vars
class ChunkedDownloadManager {
    /**
     * 
     * @param {ServiceWorker} sw 
     * @param {Uint8Array} key 
     * @param {Uint8Array} header 
     * @param {Uint8Array} nonce
     * @param {FileInfo} fileInfo
     */
    constructor(sw, key, header, nonce, fileInfo) {
        this.sw = sw;
        this.key = key;
        this.header = header;
        this.nonce = nonce;
        this.fileInfo = fileInfo;
        this.id = Math.random().toString(36).substring(2);
        this.broadcast = new BroadcastChannel(this.id);
        this.broadcast.addEventListener("message", e => this.handleBroadcastMessage(e.data));
        this.done = false;
        this.progress = 0;
    }

    handleBroadcastMessage(data) {
        console.log(data);

        // Handle messages sent by service worker
        if (data.type === "downloadAvailable") {
            if (data.id === this.id) {

                const iframe = document.createElement('iframe');
                iframe.style.display = 'none';
                iframe.src = `../../download/${this.id}`;
                document.body.appendChild(iframe);
            }
        }
    }

    error(msg) {
        // TODO: impl
    }

    /**
     * 
     * @param {ReadableStream} stream 
     */
    createDecryptionStream() {
        const bytesQueue = [];
        let state_in;
        let _done = false;
        const { header, key } = this;

        let pendingResolve;
        const waitForStream = () =>
            new Promise((resolve) => (pendingResolve = resolve));

        const decryptionStream = new ReadableStream({
            start() {
                state_in = window.sodium.crypto_secretstream_xchacha20poly1305_init_pull(header, key);
            },

            async pull(controller) {
                while (true) {
                    if (bytesQueue.length > 0) {
                        const bytes = bytesQueue.shift();
                        console.log(bytes);
                        let r1 = window.sodium.crypto_secretstream_xchacha20poly1305_pull(state_in, bytes);
                        console.log(r1);
                        controller.enqueue(r1.message);

                        if (r1.tag === 3) {
                            controller.close();
                            break;
                        }

                    } else if (_done) {
                        controller.close();
                        break;
                    } else {
                        await waitForStream();
                    }
                }
            }
        });

        return {
            stream: decryptionStream,

            /**
             * Adds a response to be decrypted by stream
             * @param {Uint8Array} response Response to be decrypted
             */
            addResponse(response) {
                bytesQueue.push(response);
                if (pendingResolve) {
                    pendingResolve();
                    pendingResolve = null;
                }
            }
        }
    }

    /**
     * Fetches first chunk based on `ENC_CHUNK_SIZE`
     * @param {number} tries how many tries left until errors
     * @returns {Promise<{encryptedFileName: Uint8Array, fileContent: Uint8Array} | undefined>}
     */
    async fetchFirstChunk(tries = 3) {
        const response = await fetch(`../../api/download/${this.fileInfo.userId}/${this.fileInfo.fileId}`, {
            headers: {
              Range: `bytes=0-${ENC_CHUNK_SIZE - 1 + 512}` // 512 for padded file name
            }
        });
        console.log("first", ENC_CHUNK_SIZE - 1 + 512);

        if (response.status === 206) {
            const data = await response.arrayBuffer();
            const uint8 = new Uint8Array(data);

            const fileNameBytes = uint8.subarray(0, 512);
            const unpaddedFileNameBytes = fileNameBytes.subarray(0, fileNameBytes.lastIndexOf(0) === -1
                ? fileNameBytes.length
                : fileNameBytes.findIndex((b, i, arr) => arr.slice(i).every(v => v === 0)));

            this.progress += uint8.length;
            return {
                encryptedFileName: unpaddedFileNameBytes,
                fileContent: uint8.subarray(512)
            };
        }

        if (tries > 0) {
            return await this.fetchFirstChunk(tries - 1)
        }

        error("Failed to fetch first chunk");
        return undefined;
    }

    /**
     * Makes request, reads and feeds to handler
     * @param {number} offset 
     * @param {(chunk: Uint8Array) => void} handler 
     */
    async fetchChunks(offset, handler) {
        const response = await fetch(`../../api/download/${this.fileInfo.userId}/${this.fileInfo.fileId}`, {
            headers: {
              Range: `bytes=${offset}-`
            }
        });
        console.log("after", offset)

        let buffer = new Uint8Array(0);
        const reader = response.body.getReader();
        while (true) {
            const { done, value } = await reader.read();
            if (done) break;

            const newBuffer = new Uint8Array(buffer.length + value.length);
            newBuffer.set(buffer, 0);
            newBuffer.set(value, buffer.length);
            buffer = newBuffer;

            while (buffer.length >= ENC_CHUNK_SIZE) {
                const chunk = buffer.slice(0, ENC_CHUNK_SIZE);
                handler(chunk);

                buffer = buffer.slice(ENC_CHUNK_SIZE);
            }
        }

        if (buffer.length > 0) {
            // Should never be called but whatever
            handler(buffer);
        }
    }

    async start() {
        const firstChunk = await this.fetchFirstChunk();
        if (!firstChunk) return;
        const { encryptedFileName, fileContent } = firstChunk;
        const fileName = sodium.to_string(sodium.crypto_secretbox_open_easy(encryptedFileName, this.nonce, this.key));

        const channel = new MessageChannel();
        this.sw.postMessage({
            type: "download",
            id: this.id,
            fileName,
            port: channel.port2
        }, [channel.port2]);

        const { stream, addResponse } = this.createDecryptionStream();
        const reader = stream.getReader();
        (async() => {
            while (true) {
                const { done, value } = await reader.read();
                if (done) {
                    channel.port1.postMessage({ done: true });
                    break;
                }

                channel.port1.postMessage({ chunk: value }, [value.buffer]);
            }
        })();

        addResponse(fileContent);
        await this.fetchChunks(ENC_CHUNK_SIZE + 512, addResponse);
    }
}

(async () => {
    await window.sodium.ready;
    ENC_CHUNK_SIZE += window.sodium.crypto_secretstream_xchacha20poly1305_ABYTES;
})();
