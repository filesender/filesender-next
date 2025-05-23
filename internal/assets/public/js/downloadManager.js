/* global sodium */
var ENC_CHUNK_SIZE = 1024 * 1024;

// eslint-disable-next-line no-unused-vars
class DownloadManager {
    /**
     * Creates a download manager instance
     * @param {Uint8Array} key 
     * @param {Uint8Array} header 
     * @param {Uint8Array} nonce
     * @param {URL} downloadUrl
     * @param {number} totalFileSize
     */
    constructor(key, header, nonce, downloadUrl, totalFileSize) {
        this.key = key;
        this.header = header;
        this.nonce = nonce;
        this.downloadUrl = downloadUrl;
        this.totalFileSize = totalFileSize;

        this.fileId = downloadUrl.pathname.split("download/")[1].split("/")[1];
        this.fileName;
        this.decryptionStream;
        this.bytesDownloaded = 0;
        this.cancelled = false;
    }

    /**
     * Sets current file name
     * @param {string} fileName 
     */
    setFileName(fileName) {
        this.fileName = fileName;
        this.bytesDownloaded += 512;
    }

    /**
     * Stops the download process
     */
    cancel() {
        this.cancelled = true;
        console.log("Received cancel signal");
    }

    /**
     * Resets the download manager
     */
    reset() {
        this.fileName = undefined;
        this.decryptionStream = undefined;
        this.bytesDownloaded = 0;
        this.cancelled = false;
    }

    /**
     * Creates a decryption stream, any chunks added will be decrypted and pushed into the resulting stream
     * @returns {{ stream: ReadableStream<Uint8Array>, addResponse: (response: Uint8Array) => void }}
     */
    createDecryptionStream() {
        const bytesQueue = [];
        const { header, key } = this;
        const state_in = window.sodium.crypto_secretstream_xchacha20poly1305_init_pull(header, key);

        let pendingResolve;
        const waitForStream = () => new Promise(resolve => pendingResolve = resolve);

        const decryptionStream = new ReadableStream({
            pull: async (controller) => {
                while (true) {
                    if (bytesQueue.length > 0) {
                        const bytes = bytesQueue.shift();
                        let r = window.sodium.crypto_secretstream_xchacha20poly1305_pull(state_in, bytes);
                        controller.enqueue(r.message);

                        if (r.tag === 3 || this.bytesDownloaded === this.totalFileSize) {
                            controller.close();
                        }
                        break;
                    } else {
                        await waitForStream();
                    }
                }
            }
        });

        this.decryptionStream = {
            stream: decryptionStream,

            /**
             * Adds a response to be decrypted by stream
             * @param {Uint8Array} response Response to be decrypted
             */
            addResponse: (response) => {
                bytesQueue.push(response);
                this.bytesDownloaded += response.length;

                if (pendingResolve) {
                    pendingResolve();
                    pendingResolve = null;
                }
            }
        }

        return this.decryptionStream;
    }

    /**
     * Creates a decryption stream if it doesn't exist already, otherwise returns already existing stream
     * @returns {{ stream: ReadableStream<Uint8Array>, addResponse: (response: Uint8Array) => void }}
     */
    getDecryptionStream() {
        if (!this.decryptionStream) {
            return this.createDecryptionStream();
        }

        return this.decryptionStream;
    }

    /**
     * Fetches the first file chunk (including file name)
     * @returns {Promise<Uint8Array>}
     */
    async fetchFirstChunk() {
        const response = await fetch(this.downloadUrl, {
            headers: {
                "Range": `bytes=0-${ENC_CHUNK_SIZE - 1 + 512}` // 512 for padded file name
            }
        });

        if (response.status !== 206) {
            const body = await response.text();
            throw new Error(`Failed fetching first chunk: ${body}`);
        }

        const data = await response.arrayBuffer();
        const uint8 = new Uint8Array(data);

        var fileNameBytes = uint8.subarray(0, 512);

        const lastIndex = fileNameBytes.lastIndexOf(0) !== -1 
            ? fileNameBytes.findIndex((b, i, arr) => arr.slice(i).every(v => v === 0))
            : fileNameBytes.length;
        fileNameBytes = fileNameBytes.subarray(0, lastIndex !== -1 ? lastIndex : fileNameBytes.length);

        this.getDecryptionStream().addResponse(uint8.subarray(512))

        return fileNameBytes;
    }

    /**
     * Fetches all chunks from offset `this.bytesDownloaded`
     */
    async fetchChunks() {
        const response = await fetch(this.downloadUrl, {
            headers: {
                "Range": `bytes=${this.bytesDownloaded}-`
            }
        });

        let buffer = new Uint8Array(0);
        const reader = response.body.getReader();
        while (!this.cancelled) {
            const { done, value } = await reader.read();
            if (done) break;

            const newBuffer = new Uint8Array(buffer.length + value.length);
            newBuffer.set(buffer, 0);
            newBuffer.set(value, buffer.length);
            buffer = newBuffer;

            while (buffer.length >= ENC_CHUNK_SIZE) {
                const chunk = buffer.slice(0, ENC_CHUNK_SIZE);
                this.getDecryptionStream().addResponse(chunk);

                buffer = buffer.slice(ENC_CHUNK_SIZE);
            }
        }

        if (this.cancelled) {
            return
        }

        if (buffer.length > 0) {
            this.getDecryptionStream().addResponse(buffer)
        }
    }

    /**
     * Starts the download, fetches first chunk & file name, adds first chunk to the decryption stream
     * @param {(fileName: string, stream: ReadableStream) => void} handler 
     */
    async start(handler) {
        await window.sodium.ready;

        const encryptedFileName = await this.fetchFirstChunk();
        const fileName = sodium.to_string(sodium.crypto_secretbox_open_easy(encryptedFileName, this.nonce, this.key))

        const { stream } = this.getDecryptionStream();
        this.setFileName(fileName);
        handler(this, fileName, stream);
    }

    /**
     * Continue the download, fetches all resuming chunks and adds them to the decryption stream
     */
    async resume() {
        if (this.bytesDownloaded === 0) throw new Error("Can't resume a download that has never started");
        let bytesDownloaded = this.bytesDownloaded.valueOf();

        while (true) {
            try {
                await this.fetchChunks();
            } catch(err) {
                console.error(`Failed downloading: ${err.message}`);
                if (bytesDownloaded !== this.bytesDownloaded) {
                    bytesDownloaded = this.bytesDownloaded.valueOf();
                    continue
                }
                throw err;
            }
            break;
        }
    }
}

(async () => {
    await window.sodium.ready;
    ENC_CHUNK_SIZE += window.sodium.crypto_secretstream_xchacha20poly1305_ABYTES;
})();
