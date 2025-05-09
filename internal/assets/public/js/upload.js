/* global sodium */
//const ENC_CHUNK_SIZE = 1024 * 1024 * 10;

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
 * Encodes bytes to base64
 * @param {Uint8Array} uint8Array Bytes to encode
 * @returns Base64 URL-safe encoded bytes
 */
const toBase64Url = (uint8Array) => {
    const base64 = btoa(String.fromCharCode(...uint8Array));
    return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

form.addEventListener("submit", async e => {
    e.preventDefault();
    setLoader(0);

    const formData = new FormData(form);
    const userId = formData.get("user-id").toString();
    const file = formData.get("file");

    if (file.name === "") {
        return showError("You have to select a file");
    }

    const fileSize = file.size;

    await window.sodium.ready;
    const key = sodium.crypto_secretstream_xchacha20poly1305_keygen();

    let nonce = sodium.randombytes_buf(sodium.crypto_secretbox_NONCEBYTES);
    const fileName = sodium.crypto_secretbox_easy(sodium.from_string(file.name), nonce, key);

    const manager = new UploadManager(userId, key, fileName);
    (async () => {
        while (true) {
            setLoader(fileSize / manager.processedBytes)
            await new Promise(resolve => setTimeout(resolve, 100));
        }
    })();

    manager.setFile(file);
    await manager.process();

    const keyEncoded = toBase64Url(key);
    const headerEncoded = toBase64Url(manager.header);
    const nonceEncoded = toBase64Url(nonce);
    
    if (manager.fileId) {
        window.location.href = `download/${userId}/${manager.fileId}#${keyEncoded}.${headerEncoded}.${nonceEncoded}`;
    }
});
