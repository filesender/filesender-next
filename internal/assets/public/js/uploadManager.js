/* global sodium */
var ENC_CHUNK_SIZE = 1024 * 1024;

// eslint-disable-next-line no-unused-vars
class UploadManager {
    /**
     * Creates an upload manager instance
     * @param {string} userId 
     */
    constructor(userId) {
        this.userId = userId;
        this.fileId;

        this.file;
        this.processedBytes = 0;
        this.uploadedBytes = 0;
        this.partialUploadLocation;
        this.state;
        this.header;
    }

    /**
     * Sets File object, encrypted file name, key & nonce used for upload
     * @param {File} file 
     * @param {Uint8Array} key
     * @param {Uint8Array} nonce
     * @param {Uint8Array} encryptedFileName
     */
    setFile(file, key, nonce, encryptedFileName) {
        this.file = file;
        this.key = key;
        this.nonce = nonce;
        this.encryptedFileName = encryptedFileName;

        this.processedBytes = 0;
        this.uploadedBytes = 0;

        const { state, header } = sodium.crypto_secretstream_xchacha20poly1305_init_push(this.key);
        this.state = state;
        this.header = header;
    }

    /**
     * Encrypts bytes using sodium
     * @param {Uint8Array} bytes 
     * @param {boolean} done
     */
    encrypt(bytes, done) {
        const encrypted = sodium.crypto_secretstream_xchacha20poly1305_push(
            this.state, bytes, null, 
            done ? sodium.crypto_secretstream_xchacha20poly1305_TAG_FINAL
                 : sodium.crypto_secretstream_xchacha20poly1305_TAG_MESSAGE
        );

        return encrypted;
    }

    /**
     * Uploads first chunk, initialises file
     * @param {Uint8Array} data 
     * @param {boolean} done 
     */
    async uploadFirstChunk(data, done) {
        const formData = new FormData();
        formData.append("file", new Blob([data]), "data.bin");

        var uploadComplete = "1";
        if (!done) {
            uploadComplete = "0";
        }

        const response = await fetch("api/upload", {
            method: "POST",
            body: formData,
            headers: {
                "Upload-Complete": uploadComplete
            }
        });

        if (response.status === 202 && !done) {
            this.uploadedBytes = data.length;
            this.partialUploadLocation = response.headers.get("Location");
            return;
        }

        if (response.status === 200) {
            const partials = response.url.split('view/')[1];
            this.fileId = partials.split("/")[1];
            return;
        }

        console.error(await response.text());
        throw new Error("Failed initialising file upload.");
    }

    /**
     * Appends file data to already existing upload
     * @param {Uint8Array} data 
     * @param {boolean} done 
     */
    async uploadChunk(data, done) {
        const formData = new FormData();
        formData.append("file", new Blob([data]), "data.bin");

        var uploadComplete = "0";
        if (done) {
            uploadComplete = "1";
        }

        const response = await fetch(this.partialUploadLocation, {
            method: "PATCH",
            body: formData,
            headers: {
                "Upload-Complete": uploadComplete,
                "Upload-Offset": this.uploadedBytes
            }
        });

        if (response.status === 202) {
            this.uploadedBytes += data.length;
            this.partialUploadLocation = response.headers.get("Location");
            return;
        }

        if (response.status === 200) {
            const partials = response.url.split('view/')[1];
            this.fileId = partials.split("/")[1];
            return;
        }

        console.error(await response.text());
        throw new Error("Failed uploading file.");
    }

    /**
     * Uploads a chunk, either initialises the upload if the upload hans't been initialised yet, or appends data to the already initialised upload
     * @param {Uint8Array} data 
     * @param {boolean} done 
     */
    async processChunk(data, done) {
        if (this.processedBytes === 0) {
            await this.uploadFirstChunk(data, done);
        } else {
            await this.uploadChunk(data, done);
        }
    }

    /**
     * If resuming upload from an errored state, we'll need to "skip" bytes from the reader
     * @param {ReadableStreamDefaultReader<Uint8Array>} reader 
     * @param {number} skip 
     */
    async skipBytes(reader, skip) {
        let totalSkipped = 0;

        while (totalSkipped < skip) {
            const { value, done } = await reader.read();
            if (done) break;

            const remainingToSkip = skip - totalSkipped;

            if (value.length <= remainingToSkip) {
                totalSkipped += value.length;
            } else {
                const remaining = value.slice(remainingToSkip);
                return {
                    reader,
                    remaining
                }
            }
        }

        return {
            reader,
            remaining: null
        }
    }

    /**
     * Creates file reader, skips bytes if the upload was already started before
     */
    async createReader() {
        let reader = this.file.stream().getReader();
        let buffer = new Uint8Array(0);

        if (this.processedBytes > 0) {
            const { reader: newReader, remaining } = await this.skipBytes(reader, this.processedBytes);
            return { reader: newReader, buffer: remaining };
        }

        return { reader, buffer };
    }

    /**
     * Starts processing file, chunks, encrypts & uploads
     */
    async process() {
        if (!this.file) return;

        let { reader, buffer } = await this.createReader();
        let chunkSize = 0;
        let doneReading = false;

        while (true) {
            while (buffer.length <= ENC_CHUNK_SIZE && !doneReading) {
                const { value, done } = await reader.read();
                if (done) {
                    doneReading = true;
                    break;
                }
    
                const tmp = new Uint8Array(buffer.length + value.length);
                tmp.set(buffer, 0);
                tmp.set(value, buffer.length);
                buffer = tmp;
            }

            let encrypted;
            if (buffer.length >= ENC_CHUNK_SIZE) {
                const chunk = buffer.slice(0, ENC_CHUNK_SIZE);
                buffer = buffer.slice(ENC_CHUNK_SIZE);

                chunkSize = chunk.length;
                encrypted = this.encrypt(chunk, false);
            } else if (doneReading && buffer.length > 0) {
                chunkSize = buffer.length;
                encrypted = this.encrypt(buffer, true);
            }

            if (this.processedBytes === 0) {
                const combined = new Uint8Array(512 + encrypted.length);
                combined.set(this.encryptedFileName.subarray(0, 512));
                combined.set(encrypted, 512);

                chunkSize += 512;
                encrypted = combined;
            }

            await this.processChunk(encrypted, doneReading);
            this.processedBytes += chunkSize;

            if (doneReading) {
                break;
            }
        }
    }
}
