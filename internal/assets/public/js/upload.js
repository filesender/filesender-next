/* global sodium, showError, hideError, setProgress */
const form = document.querySelector("form");
const fileSelector = document.querySelector("#files-selector");

/**
 * Encodes bytes to base64
 * @param {Uint8Array} uint8Array Bytes to encode
 * @returns Base64 URL-safe encoded bytes
 */
const toBase64Url = (uint8Array) => {
    const base64 = btoa(String.fromCharCode(...uint8Array));
    return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

// eslint-disable-next-line no-undef
const manager = new UploadManager();

fileSelector.addEventListener("change", () => {
    const formData = new FormData(form);
    const file = formData.get("file");

    if (file !== manager.file) {
        form.querySelector('input[type="submit"]').value = "Upload";
    }
});

form.addEventListener("submit", async e => {
    e.preventDefault();
    hideError();

    const formData = new FormData(form);
    const file = formData.get("file");

    if (file.name === "") {
        return showError("You have to select a file");
    }

    form.querySelector('input[type="submit"]').disabled = true;

    await window.sodium.ready;
    if (file !== manager.file) {
        const key = sodium.crypto_secretstream_xchacha20poly1305_keygen();
        let nonce = sodium.randombytes_buf(sodium.crypto_secretbox_NONCEBYTES);
        const fileName = sodium.crypto_secretbox_easy(sodium.from_string(file.name), nonce, key);
        manager.setFile(file, key, nonce, fileName);
    }

    (async () => {
        const max = file.size;
        while (true) {
            setProgress(manager.processedBytes, max);
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

            tries++;
        }
    }

    if (done) {
        const keyEncoded = toBase64Url(manager.key);
        const headerEncoded = toBase64Url(manager.header);
        const nonceEncoded = toBase64Url(manager.nonce);
        
        if (manager.downloadLink) {
            window.location.href = `${manager.downloadLink}#${keyEncoded}.${headerEncoded}.${nonceEncoded}`;
        }
    } else {
        form.querySelector('input[type="submit"]').value = "Resume Upload";
        showError(`Failed uploading file: ${err.message}`)
    }

    form.querySelector('input[type="submit"]').disabled = false;
});
