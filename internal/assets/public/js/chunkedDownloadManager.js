/* global sodium */
var ENC_CHUNK_SIZE = 1024 * 1024 * 10;

// eslint-disable-next-line no-unused-vars
class ChunkedDownloadManager {
    /**
     * 
     * @param {Uint8Array} key 
     * @param {Uint8Array} header 
     * @param {Uint8Array} nonce
     * @param {string} userId
     * @param {string} fileId
     */
    constructor(key, header, nonce, userId, fileId) {
        this.key = key;
        this.header = header;
        this.nonce = nonce;
        this.userId = userId;
        this.fileId = fileId;
        this.done = false;
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
                        let r1 = window.sodium.crypto_secretstream_xchacha20poly1305_pull(state_in, bytes);
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
        const response = await fetch(`../../api/download/${this.userId}/${this.fileId}`, {
            headers: {
              Range: `bytes=0-${ENC_CHUNK_SIZE - 1 + 512}` // 512 for padded file name
            }
        });

        if (response.status === 206) {
            const data = await response.arrayBuffer();
            console.log(data.byteLength);
            const uint8 = new Uint8Array(data);

            const fileNameBytes = uint8.subarray(0, 512);
            const unpaddedFileNameBytes = fileNameBytes.subarray(0, fileNameBytes.lastIndexOf(0) === -1
                ? fileNameBytes.length
                : fileNameBytes.findIndex((b, i, arr) => arr.slice(i).every(v => v === 0)));

            return {
                encryptedFileName: unpaddedFileNameBytes,
                fileContent: uint8.subarray(512)
            };
        }

        if (tries > 0) {
            return await this.fetchFirstChunk(tries - 1)
        }

        return undefined;
    }

    /**
     * Makes request, reads and feeds to handler
     * @param {number} offset 
     * @param {(chunk: Uint8Array) => void} handler 
     */
    async fetchChunks(offset, handler) {
        const response = await fetch(`../../api/download/${this.userId}/${this.fileId}`, {
            headers: {
              Range: `bytes=${offset}-`
            }
        });

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

    async getTotalSize() {
        const response = await fetch(`../../api/download/${this.userId}/${this.fileId}`, {
            method: "HEAD"
        });

        return parseInt(response.headers.get("content-length"))
    }

    /**
     * 
     * @param {(fileName: string, stream: ReadableStream) => void} handler 
     * @returns 
     */
    async start(handler) {
        await window.sodium.ready;

        const firstChunk = await this.fetchFirstChunk();
        if (!firstChunk) return;
        const { encryptedFileName, fileContent } = firstChunk;
        const fileName = sodium.to_string(sodium.crypto_secretbox_open_easy(encryptedFileName, this.nonce, this.key));

        const { stream, addResponse } = this.createDecryptionStream();
        handler(fileName, stream);

        addResponse(fileContent);
        await this.fetchChunks(ENC_CHUNK_SIZE + 512, addResponse);
    }
}

(async () => {
    await window.sodium.ready;
    ENC_CHUNK_SIZE += window.sodium.crypto_secretstream_xchacha20poly1305_ABYTES;
})();
