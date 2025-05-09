/* global sodium */
const ENC_CHUNK_SIZE = 1024 * 1024 * 10;

const form = document.querySelector("form");
var userId = "";
var partialUploadLocation = "";

const setLoader = (progress) => {
    if (progress > 1) {
        progress = 1;
    }

    const loader = document.querySelector("div.loader");
    loader.style.width = `${progress * 100}%`;

    const loaderText = document.querySelector("p#progress");
    loaderText.innerText = `${Math.round(progress * 10000) / 100}%`;
}

/**
 * Dummy error handling function
 * @param {string} msg Message to show to use
 */
const showError = msg => {
    console.log(`ERROR: ${msg}`);
}

/**
 * Calculate expected size of encrypted file using libsodium secretstream
 * @param {number} fileSize - The original file size in bytes
 * @returns {number} - Expected encrypted file size in bytes
 */
function calculateEncryptedSize(fileSize) {
    const headerSize = sodium.crypto_secretstream_xchacha20poly1305_HEADERBYTES;
    const chunkOverhead = sodium.crypto_secretstream_xchacha20poly1305_ABYTES;
  
    const chunkCount = Math.ceil(fileSize / ENC_CHUNK_SIZE);
    const encryptedSize = fileSize + headerSize + chunkCount * chunkOverhead;
  
    return encryptedSize;
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
 * @param {Uint8Array} data 
 * @param {boolean} partial If the file being uploaded is chunked or not
 * @param {Uint8Array} fileName Encrypted file name
 * @returns `false` if errored, `string` if last chunk & successful, `true` if successful
 */
const uploadFile = async (data, partial, fileName) => {
    const combined = new Uint8Array(512 + data.length);
    combined.set(fileName.subarray(0, 512));
    combined.set(data, 512);

    const formData = new FormData();
    formData.append("file", new Blob([combined]), "data.bin");

    var uploadComplete = "?1";
    if (partial) {
        uploadComplete = "?0"
    }
    
    const response = await fetch("api/upload", {
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
 * @param {Uint8Array} file The chunk as a `File` object
 * @param {number} offset Byte offset of the chunk
 * @param {boolean} done If this is the last chunk
 * @returns `false` if errored, `string` if last chunk & successful, `true` if successful
 */
const uploadPartialFile = async (file, offset, done) => {
    const formData = new FormData();
    formData.append("file", new Blob([file]), "data.bin");

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
    setLoader(0);

    await window.sodium.ready;
    const sodium = window.sodium;

    const formData = new FormData(form);
    const file = formData.get("file");

    if (file.name === "") {
        return showError("You have to select a file");
    }

    const expectedFileSize = calculateEncryptedSize(file.size);

    let key = sodium.crypto_secretstream_xchacha20poly1305_keygen();
    let [stream, header] = createEncryptedStream(file, key);

    let nonce = sodium.randombytes_buf(sodium.crypto_secretbox_NONCEBYTES);
    const fileName = sodium.crypto_secretbox_easy(sodium.from_string(file.name), nonce, key);

    console.log("Key", key);
    console.log("Header", header);

    let fileId = false;
    let total = 0;
    let i = 0;
    
    const reader = stream.getReader();
    let { value, done } = await reader.read();
    
    while (!done) {
        const nextChunk = await reader.read();
        const moreComing = !nextChunk.done;

        if (total === 0) {
            const res = await uploadFile(value, moreComing, fileName);
            if (res === false) return;
    
            total += value.length + 512; // 512 bytes filename prefix

            if (res !== true) {
                fileId = res;
                break;
            }
        } else {
            const res = await uploadPartialFile(value, total, !moreComing);
            if (res === false) return;
    
            total += value.length;
    
            if (res !== true) {
                fileId = res;
                break;
            }
        }

        setLoader(total/expectedFileSize);
    
        // Move to next chunk
        value = nextChunk.value;
        done = nextChunk.done;
        i++;
    }
    setLoader(1);
    
    const keyEncoded = toBase64Url(key);
    const headerEncoded = toBase64Url(header);
    const nonceEncoded = toBase64Url(nonce);
    
    if (fileId !== false) {
        window.location.href = `download/${userId}/${fileId}#${keyEncoded}.${headerEncoded}.${nonceEncoded}`;
    }
});
