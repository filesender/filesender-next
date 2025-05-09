/* global sodium */
const errorBox = document.querySelector("div.error");
const form = document.querySelector("form");

const setLoader = (progress) => {
    if (!progress) {
        const loader = document.querySelector("div.loader");
        loader.style.width = "0%";

        const loaderText = document.querySelector("p#progress");
        loaderText.innerText = "";
    }

    if (progress > 1) {
        progress = 1;
    }

    const loader = document.querySelector("div.loader");
    loader.style.width = `${progress * 100}%`;

    const loaderText = document.querySelector("p#progress");
    loaderText.innerText = `${Math.round(progress * 10000) / 100}%`;
}

/**
 * Error showing
 * @param {string} msg Message to show to use
 */
const showError = msg => {
    console.log(`ERROR: ${msg}`);
    errorBox.innerText = msg;
    errorBox.classList.remove("hidden");
}

/**
 * Hides whatever current error message is being shown
 */
const hideError = () => {
    errorBox.classList.add("hidden");
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

const formData = new FormData(form);
const userId = formData.get("user-id").toString();
// eslint-disable-next-line no-undef
const manager = new UploadManager(userId);

form.addEventListener("submit", async e => {
    e.preventDefault();
    hideError();

    const formData = new FormData(form);
    const file = formData.get("file");

    if (file.name === "") {
        return showError("You have to select a file");
    }

    await window.sodium.ready;
    const fileSize = file.size;
    if (file !== manager.file) {
        const key = sodium.crypto_secretstream_xchacha20poly1305_keygen();
        let nonce = sodium.randombytes_buf(sodium.crypto_secretbox_NONCEBYTES);
        const fileName = sodium.crypto_secretbox_easy(sodium.from_string(file.name), nonce, key);
        manager.setFile(file, key, nonce, fileName);
    }

    (async () => {
        while (true) {
            setLoader(fileSize / manager.processedBytes)
            await new Promise(resolve => setTimeout(resolve, 100));
        }
    })();

    let done = false;
    let tries = 0;
    let err;
    while (tries < 3) {
        try {
            hideError();
            await manager.process();
            done = true;
            break;
        } catch(e) {
            console.error(e);
            err = e;

            let message = `Failed uploading file: ${e.message}.`;
            if (tries < 3) message += " Retrying in 5 seconds."
            showError(message);

            if (tries < 3) await new Promise(resolve => setTimeout(resolve, 5000));
            else break;
        }
    }

    if (done) {
        const keyEncoded = toBase64Url(manager.key);
        const headerEncoded = toBase64Url(manager.header);
        const nonceEncoded = toBase64Url(manager.nonce);
        
        if (manager.fileId) {
            window.location.href = `download/${userId}/${manager.fileId}#${keyEncoded}.${headerEncoded}.${nonceEncoded}`;
        }
    } else {
        form.querySelector('input[type="submit"]').innerText = "Continue";
        showError(`Failed uploading file: ${err.message}`)
    }
});
