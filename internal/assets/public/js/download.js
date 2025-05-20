/* global createServiceWorkerHandler, createMemoryHandler, createFileSystemHandler */
const errorBox = document.querySelector("div.error");
const progress = document.querySelector("progress");
const a = document.querySelector("a");
const button = a.children[0];
const downloadUrl = new URL(a.href);

const isSaveFilePickerSupported = "showSaveFilePicker" in window;

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

const getHashVars = () => {
    const [key, header, nonce] = window.location.hash.substring(1).split(".").map(v => fromBase64Url(v));
    if (!key || !header || !nonce) {
        return [false, false, false];
    }

    if (key.length !== 32 || header.length !== 24 || header.length !== 24) {
        return [false, false, false];
    }

    return [key, header, nonce];
}

if (parseInt(progress.max) > 1024 * 1024 * 500 && !isSaveFilePickerSupported) {
    (async () => {
        const parts = window.location.pathname.split("/view");
        parts.pop();

        await navigator.serviceWorker.register("../../js/sw.js", { scope: `${parts.join("/view")}/` }).catch(err => {
            console.error(err);
            showError(`Failed registering service worker: ${err.message}`);
        });
    })();
}

const [key, header, nonce] = getHashVars();
if (!key) {
    showError("No key, header, or nonce present in url!");
    a.remove();
} else {
    console.log(key, header, nonce)
    // eslint-disable-next-line no-undef
    const manager = new DownloadManager(key, header, nonce, downloadUrl, parseInt(progress.max));

    (async () => {
        while (true) {
            progress.value = manager.bytesDownloaded;

            if (manager.bytesDownloaded > 0) {
                const loaderText = document.querySelector("p#progress");
                loaderText.innerText = `${(Math.round(progress.value / progress.max * 10000) / 100).toFixed(2)}%`;
            }

            await new Promise(resolve => setTimeout(resolve, 100));
        }
    })();

    a.addEventListener("click", async e => {
        e.preventDefault();
        hideError();
    
        await window.sodium.ready;
        button.disabled = true;

        if (manager.cancelled) {
            // In case of service worker handler, we want to reset manager when cancelled (we can't continue after cancel)
            manager.reset();
        }

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
                    handler = createServiceWorkerHandler(ready, manager.fileId, sw.active);
                }

                await manager.start(handler, showError);
                await waitForHandler;
            } catch (err) {
                console.error(err);
                showError(`Failed to start download: ${err.message}`);
                button.disabled = false;
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

            let message = `Failed downloading file: ${err.message}.`;
            if (tries < 3) message += " Retrying in 5 seconds.";
            showError(message);

            if (tries < 3) await new Promise(resolve => setTimeout(resolve, 5000));
        }

        if (manager.bytesDownloaded !== manager.totalFileSize && err) {
            button.innerText = "Resume Download";
        }

        button.disabled = false;
    });
}
