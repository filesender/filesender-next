/* global createServiceWorkerHandler */
const errorBox = document.querySelector("div.error");
const form = document.querySelector("form");

const setLoader = (progress) => {
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
 * Decodes base64 string to bytes
 * @param {string} base64url 
 * @returns bytes
 */
const fromBase64Url = (base64url) => {
    const base64 = base64url.replace(/-/g, '+').replace(/_/g, '/').padEnd(base64url.length + (4 - base64url.length % 4) % 4, '=');

    const binary = atob(base64);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i++) {
        bytes[i] = binary.charCodeAt(i);
    }
    return bytes;
}

const isSaveFilePickerSupported = "showSaveFilePicker" in window;

const formData = new FormData(form);
const userId = formData.get("user-id").toString();
const fileId = formData.get("file-id").toString();
const byteSize = formData.get("byte-size").toString();

if (byteSize <= 1024 * 1024 * 500 || !isSaveFilePickerSupported) {
    (async () => {
        await navigator.serviceWorker.register("../../sw.js").catch(err => {
            console.error(err);
            showError(`Failed registering service worker: ${err.message}`);
        });
    })();
}

const [key, header, nonce] = window.location.hash.substring(1).split(".").map(v => fromBase64Url(v));
if (!key || !header || !nonce) {
    showError("No key, header, or nonce present in url!");
} else {
    // eslint-disable-next-line no-undef
    const manager = new DownloadManager(key, header, nonce, userId, fileId, byteSize);

    (async () => {
        while (true) {
            if (manager.bytesDownloaded !== 0) {
                const progress = manager.bytesDownloaded / manager.totalFileSize;
                setLoader(progress);

                if (progress >= 1) {
                    break;
                }
            }

            await new Promise(resolve => setTimeout(resolve, 100));
        }
    })();

    form.addEventListener("submit", async e => {
        e.preventDefault();
        hideError();
    
        await window.sodium.ready;
        form.querySelector("button").disabled = true;

        if (manager.bytesDownloaded === 0) {
            try {
                let handler, ready;
                const waitForHandler = new Promise(resolve => ready = resolve)

                if (manager.totalFileSize <= 1024 * 1024 * 500) { // 500MB
                    console.log("Using memory handler");
                    handler = createMemoryHandler(ready);
                } else if (isSaveFilePickerSupported) {
                    console.log("Using file system API hanlder");
                    handler = createFileSystemHandler(ready);
                } else {
                    console.log("Using service worker handler");
                    const sw = await navigator.serviceWorker.ready;
                    handler = createServiceWorkerHandler(ready, fileId, sw.active);
                }

                await manager.start(handler, showError);
                await waitForHandler;
            } catch (err) {
                console.error(err);
                showError(`Failed to start download: ${err.message}`);
                form.querySelector("button").disabled = false;
                return;
            }
        }

        let tries = 0;
        let err;
        while (tries < 3) {
            try {
                await manager.resume();
                err = undefined;
                break;
            } catch (e) {
                console.error(e);
                err = e;
            }

            tries++;

            let message = `Failed uploading file: ${err.message}.`;
            if (tries < 3) message += " Retrying in 5 seconds.";
            showError(message);

            if (tries < 3) await new Promise(resolve => setTimeout(resolve, 5000));
        }

        if (manager.bytesDownloaded !== manager.totalFileSize && err) {
            form.querySelector("button").innerText = "Resume Download";
        }

        form.querySelector("button").disabled = false;
    });
}
