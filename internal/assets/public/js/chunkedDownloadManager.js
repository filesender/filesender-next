
/**
 * @typedef {Object} FileInfo
 * @property {boolean} available
 * @property {boolean} chunked
 * @property {number} chunkCount
 * @property {string} userId
 * @property {string} fileId
 * @property {string} fileName
 */

class ChunkedDownloadManager {
    /**
     * 
     * @param {ServiceWorker} sw 
     * @param {Uint8Array} key 
     * @param {Uint8Array} header 
     * @param {FileInfo} fileInfo
     */
    constructor(sw, key, header, fileInfo) {
        this.sw = sw;
        this.key = key;
        this.header = header;
        this.fileInfo = fileInfo;
        this.id = Math.random().toString(36).substring(2);
        this.broadcast = new BroadcastChannel(this.id);
        this.broadcast.addEventListener("message", e => this.handleBroadcastMessage(e.data));
        this.done = false;
        this.progress = 0;

        if (!this.fileInfo.chunked) {
            this.chunks = [`../../api/v1/download/${fileInfo.userId}/${fileInfo.fileId}/0`]
        } else {
            this.chunks = Array.from({ length: this.fileInfo.chunkCount+1 }, (_, i) => `../../api/v1/download/${fileInfo.userId}/${fileInfo.fileId}/${i}`)
        }

        this.total = this.chunks.length;
    }

    /**
     * 
     * @param {string} url 
     * @returns Body reader
     */
    async downloadChunk(url) {
        const response = await fetch(url);
        this.progress += 1;
        return await response.bytes()
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
            start(_controller) {
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

    async start() {
        const channel = new MessageChannel();
        this.sw.postMessage({
            type: "download",
            id: this.id,
            fileName: this.fileInfo.fileName,
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

        let buffer = new Uint8Array(0);
        const response = await fetch(`../../api/v1/download/${this.fileInfo.userId}/${this.fileInfo.fileId}`);
        const responseReader = response.body.getReader();
        while (true) {
            const { done, value } = await responseReader.read();
            if (done) break;
            this.progress += value.length;

            const newBuffer = new Uint8Array(buffer.length + value.length);
            newBuffer.set(buffer, 0);
            newBuffer.set(value, buffer.length);
            buffer = newBuffer;

            while (buffer.length >= this.fileInfo.chunkSize) {
                const chunk = buffer.slice(0, this.fileInfo.chunkSize);
                addResponse(chunk);

                buffer = buffer.slice(this.fileInfo.chunkSize);
            }
        }

        if (buffer.length > 0) {
            addResponse(buffer);
        }
    }
}
