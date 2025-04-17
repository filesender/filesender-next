const ENC_CHUNK_SIZE = 1024 * 1024;

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
 * Encodes bytes to base64
 * @param {Uint8Array} uint8Array Bytes to encode
 * @returns Base64 URL-safe encoded bytes
 */
const toBase64Url = (uint8Array) => {
    const base64 = btoa(String.fromCharCode(...uint8Array));
    return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

/**
 * Uploads a file
 * @param {string} expiryDate `YYYY-MM-DD` formatted expiry date of the file
 * @param {File} file 
 * @param {boolean} partial
 * @returns `false` if errored, `string` if last chunk & successful, `true` if successful
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

/**
 * Uploads a file chunk
 * @param {File} file The chunk as a `File` object
 * @param {number} offset Byte offset of the chunk
 * @param {boolean} done If this is the last chunk
 * @returns `false` if errored, `string` if last chunk & successful, `true` if successful
 */
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
 * Keeps a maximum of 10 chunks in memory (10 x `ENC_CHUNK_SIZE`)
 * @param {File} f 
 * @param {Uint8Array} key
 * @returns {[ReadableStream<Uint8Array>, Uint8Array]}
 */
const createEncryptedStream = (f, key) => {
    const { state, header } = sodium.crypto_secretstream_xchacha20poly1305_init_push(key);
    const reader = f.stream().getReader();

    let buffer = new Uint8Array(0);
    let doneReading = false;

    const encryptedStream = new ReadableStream({
        async pull(controller) {
            while (buffer.byteLength < ENC_CHUNK_SIZE && !doneReading) {
                const { value, done } = await reader.read();
                if (done) {
                    doneReading = true;
                    break;
                }

                // Append new chunk to buffer
                const tmp = new Uint8Array(buffer.length + value.length);
                tmp.set(buffer, 0);
                tmp.set(value, buffer.length);
                buffer = tmp;
            }

            if (buffer.length >= ENC_CHUNK_SIZE) {
                const chunk = buffer.slice(0, ENC_CHUNK_SIZE);
                buffer = buffer.slice(ENC_CHUNK_SIZE);

                const encrypted = sodium.crypto_secretstream_xchacha20poly1305_push(
                    state, chunk, null, sodium.crypto_secretstream_xchacha20poly1305_TAG_MESSAGE
                );
                controller.enqueue(encrypted);
            } else if (doneReading && buffer.length > 0) {
                const encrypted = sodium.crypto_secretstream_xchacha20poly1305_push(
                    state, buffer, null, sodium.crypto_secretstream_xchacha20poly1305_TAG_FINAL
                );
                controller.enqueue(encrypted);
                controller.close();
            } else if (doneReading) {
                // No remaining data to encrypt, just close
                controller.close();
            }
        }
    }, {
        highWaterMark: ENC_CHUNK_SIZE * 10,
        size(chunk) {
            return chunk.byteLength;
        }
    });

    return [encryptedStream, header];
};

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

    let fileId = false;
    let total = 0;
    let i = 0;

    const reader = stream.getReader();
    var { value, done } = await reader.read();
    while (true) {
        const blob = new Blob(value);
        const file = new File([blob], `${i}.bin`);

        if (total === 0) {
            const res = await uploadFile(expiryDate, file, !done);
            if (res === false) return;

            total += file.size;
    
            if (res !== true) {
                fileId = res;
                break;
            }
        } else {
            const res = await uploadPartialFile(file, total, done);
            if (res === false) return;

            total += file.size;
    
            if (res !== true) {
                fileId = res;
                break;
            }
        }

        if (done) break;
        const res = await reader.read();
        value = res.value;
        done = res.done;
        i++;
    }

    const keyEncoded = toBase64Url(key);
    const headerEncoded = toBase64Url(header);

    if (fileId !== false) {
        window.location.replace(`download/${userId}/${fileId}#${keyEncoded}.${headerEncoded}`);
    }

});
