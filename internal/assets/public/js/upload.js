const CHUNK_SIZE = 1024 * 1024;

const form = document.querySelector("form");
var userId = "";

/**
 * Dummy error handling function
 * @param {string} msg Message to show to use
 */
const showError = msg => {
    console.log(`ERROR: ${msg}`);
}

/**
 * Uploads a file
 * @param {string} expiryDate `YYYY-MM-DD` formatted expiry date of the file
 * @param {File} file 
 * @returns {Promise<string|false>} Contains file ID with successful, otherwise `false`
 */
const uploadFile = async (expiryDate, file) => {
    const formData = new FormData();
    formData.append("file", file);
    formData.append("expiry-date", expiryDate);
    
    const response = await fetch("api/v1/upload", {
        method: "POST",
        body: formData
    });

    if (response.status === 200) {
        const parials = response.url.split('download/')[1];
        return parials.split("/")[1];
    }

    showError("Something went wrong uploading file");
    console.error(response.body)
    return false;
}

/**
 * Turns selected file into an encrypted readable byte stream.
 * Keeps a maximum of 10 chunks in memory (10 x `CHUNK_SIZE`)
 * @param {File} f 
 * @param {Uint8Array} key
 * @returns {[ReadableStream<Uint8Array>, Uint8Array]}
 */
const createEncryptedStream = (f, key) => {
    let res = sodium.crypto_secretstream_xchacha20poly1305_init_push(key);
    let [state_out, header] = [res.state, res.header];

    const fileReader = f.stream().getReader();

    let prevChunk = null;
    const encryptedStream = new ReadableStream({
        async start(_controller) {
            const { value } = await fileReader.read();
            prevChunk = value;
        },

        async pull(controller) {
            const { done, value } = await fileReader.read();
            if (done) {
                if (prevChunk) {
                    let msg = sodium.crypto_secretstream_xchacha20poly1305_push(state_out, prevChunk, null, sodium.crypto_secretstream_xchacha20poly1305_TAG_FINAL);
                    controller.enqueue(msg);
                }

                controller.close();
                return;
            }

            let msg = sodium.crypto_secretstream_xchacha20poly1305_push(state_out, prevChunk, null, sodium.crypto_secretstream_xchacha20poly1305_TAG_MESSAGE);
            controller.enqueue(msg);
        
            prevChunk = value;
        }
    }, {
        highWaterMark: CHUNK_SIZE * 10,
        size(chunk) {
            return chunk.byteLength;
        }
    });

    return [
        encryptedStream,
        header
    ]
}

/**
 * Takes readable byte stream, creates a `File` generator in chunks of `CHUNK_SIZE`
 * @param {ReadableStream<Uint8Array>} stream 
 * @param {number} chunkSize 
 */
async function* fileChunkGenerator(stream, chunkSize) {
    const reader = stream.getReader();
    let chunks = [];
    let accumulatedSize = 0;
    let fileIndex = 0;

    while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        chunks.push(value);
        accumulatedSize += value.length;

        while (accumulatedSize >= chunkSize) {
            let sizeToUse = 0;
            const tempChunks = [];

            while (sizeToUse < chunkSize && chunks.length > 0) {
                const chunk = chunks.shift();
                const remaining = chunkSize - sizeToUse;

                if (chunk.length <= remaining) {
                    tempChunks.push(chunk);
                    sizeToUse += chunk.length;
                    accumulatedSize -= chunk.length;
                } else {
                    tempChunks.push(chunk.slice(0, remaining));
                    chunks.unshift(chunk.slice(remaining));
                    accumulatedSize -= remaining;
                    sizeToUse += remaining;
                }
            }

            const blob = new Blob(tempChunks);
            const file = new File([blob], `${fileIndex}.bin`);
            fileIndex++;
            yield file;
        }
    }

    if (accumulatedSize > 0) {
        const blob = new Blob(chunks);
        const file = new File([blob], `${fileIndex}.bin`);
        yield file;
    }
}

form.addEventListener("submit", async e => {
    e.preventDefault();
    await window.sodium.ready;
    const sodium = window.sodium;

    const formData = new FormData(form);
    const file = formData.get("file");
    const expiryDate = formData.get("expiry-date");

    if (file.name === "") {
        return showError("You have to select a file");
    }

    let key = sodium.crypto_secretstream_xchacha20poly1305_keygen();
    let [stream, header] = createEncryptedStream(file, key);

    console.log("Key", key);
    console.log("Header", header);

    for await (const file of fileChunkGenerator(stream, CHUNK_SIZE)) {
        console.log(file);
    }
});
