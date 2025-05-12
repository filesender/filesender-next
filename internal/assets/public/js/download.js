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

(async () => {
    await navigator.serviceWorker.register("../../sw.js").catch(err => {
        console.error(err);
        showError(`Failed registering service worker: ${err.message}`);
    });
})();

/**
 * Hides whatever current error message is being shown
 */
const hideError = () => {
    errorBox.classList.add("hidden");
}

/**
 * 
 * @param {string | undefined} msg 
 */
const errorMessageHandler = (msg) => {
    if (msg) {
        showError(msg);
    } else {
        hideError();
    }
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

const formData = new FormData(form);
const userId = formData.get("user-id").toString();
const fileId = formData.get("file-id").toString();

const [key, header, nonce] = window.location.hash.substring(1).split(".").map(v => fromBase64Url(v));
if (!key || !header || !nonce) {
    showError("No key, header, or nonce present in url!");
} else {
    console.log("Key", key)
    console.log("Header", header);
    console.log("Nonce", nonce);

    // eslint-disable-next-line no-undef
    const manager = new DownloadManager(key, header, nonce, userId, fileId);

    form.addEventListener("submit", async e => {
        e.preventDefault();
        hideError();
    
        await window.sodium.ready;
        const totalFileSize = await manager.getTotalSize();
    
        let handler;
        const sw = await navigator.serviceWorker.ready;
        handler = await createServiceWorkerHandler(sw.active, fileId);

        /**
         * 
         * @param {(fileName: string, stream: ReadableStream) => void} handler 
         * @param {(msg: string | undefined) => void} errorMessageHandler
         * @returns 
         */
        const startOrResumeDownload = async (handler, errorMessageHandler) => {
            console.log("Downloaded:", manager.bytesDownloaded);
            if (manager.bytesDownloaded === 0) {
                try {
                    await manager.start(handler, errorMessageHandler);
                    await manager.resume(errorMessageHandler);
                } catch (err) {
                    if (manager.bytesDownloaded > 0) {
                        form.querySelector("button").innerText = "Resume Download";
                    }

                    throw err;
                }
                return;
            }

            await manager.resume(errorMessageHandler);
        }
    
        (async () => {
            while (true) {
                const progress = manager.bytesDownloaded / totalFileSize;
                setLoader(progress);
                await new Promise(resolve => setTimeout(resolve, 100));
        
                if (progress >= 1) {
                    break;
                }
            }
        })();

        form.querySelector("button").disabled = true;
        
        let tries = 0;
        let err;
        while (tries < 3) {
            try {
                await startOrResumeDownload(handler, errorMessageHandler);
                break;
            } catch (e) {
                err = e;
            }
            tries++;
        }
        
        if (err) {
            showError(`Failed downloading: ${err.message}`);
        }

        form.querySelector("button").disabled = false;
    });
}
