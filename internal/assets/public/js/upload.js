const CHUNK_SIZE = 1024 * 1024;

const form = document.querySelector("form");
var userId = "";
var partialUploadLocation = "";

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
 * @param {boolean} partial
 * @returns {Promise<string|false>} Contains file ID with successful, otherwise `false`
 */
const uploadFile = async (expiryDate, file, partial) => {
    const formData = new FormData();
    formData.append("file", file);
    formData.append("expiry-date", expiryDate);

    var uploadComplete = "?1";
    if (partial) {
        uploadComplete = "?0"
    }
    
    const response = await fetch("api/v1/upload", {
        method: "POST",
        body: formData,
        headers: {
            "Upload-Complete": uploadComplete
        }
    });

    if (response.status === 202 && partial) {
        partialUploadLocation = response.headers.get("Location");
        return true;
    }

    if (response.status === 200) {
        const parials = response.url.split('download/')[1];
        return parials.split("/")[1];
    }

    showError("Something went wrong uploading file");
    console.error(response.body)
    return false;
}

const uploadPartialFile = async (file, offset, done) => {
    const formData = new FormData();
    formData.append("file", file);

    var uploadComplete = "?0";
    if (done) {
        uploadComplete = "?1"
    }

    const response = await fetch(partialUploadLocation, {
        method: "PATCH",
        body: formData,
        headers: {
            "Upload-Complete": uploadComplete,
            "Upload-Offset": offset
        }
    });

    if (response.status === 202) {
        partialUploadLocation = response.headers.get("Location");
        return true;
    }

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

    if (accumulatedSize >= 0) {
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

    let total = 0;
    let fileId = false;
    const chunkIterator = fileChunkGenerator(stream, CHUNK_SIZE);

    let first = true;
    let result = await chunkIterator.next();
    while (!result.done || first) {
        first = false;
        const file = result.value;

        const nextResult = await chunkIterator.next();
        const isLastChunk = nextResult.done;

        if (total === 0) {
            const res = await uploadFile(expiryDate, file, !isLastChunk);
            console.log(res);
            if (res === false) return;

            total += file.size;
    
            if (res !== true) {
                fileId = res;
                break;
            }
        } else {
            const res = await uploadPartialFile(file, total, isLastChunk);
            if (res === false) return;

            total += file.size;
    
            if (res !== true) {
                fileId = res;
                break;
            }
        }
    
        result = nextResult;
    }

    if (fileId !== false) {
        console.log(userId, fileId)
        window.location.replace(`download/${userId}/${fileId}`);
    }

});
