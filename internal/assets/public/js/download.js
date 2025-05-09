
const form = document.querySelector("form");

const setLoader = (progress) => {
    const loader = document.querySelector("div.loader");
    loader.style.width = `${progress * 100}%`;

    const loaderText = document.querySelector("p#progress");
    loaderText.innerText = `${Math.round(progress * 10000) / 100}%`;
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

/**
 * 
 * @param {ServiceWorker} sw
 * @param {string} fileId
 */
const createServiceWorkerHandler = async (sw, fileId) => {
    // Register service worker
    await navigator.serviceWorker.register("../../sw.js");

    /**
     * 
     * @param {string} fileName 
     * @param {ReadableStream} stream 
     */
    const handler = (fileName, stream) => {
        const broadcast = new BroadcastChannel(fileId);
        broadcast.addEventListener("message", e => {
            console.log(e.data);

            if (e.data.type === "downloadAvailable") {
                if (e.data.id === fileId) {

                    const iframe = document.createElement('iframe');
                    iframe.style.display = 'none';
                    iframe.src = `../../download/${fileId}`;
                    document.body.appendChild(iframe);
                }
            }
        });

        const channel = new MessageChannel();
        sw.postMessage({
            type: "download",
            id: fileId,
            fileName,
            port: channel.port2
        }, [channel.port2]);

        const reader = stream.getReader();
        (async() => {
            while (true) {
                const { done, value } = await reader.read();
                if (done) {
                    channel.port1.postMessage({ done: true });
                    break;
                }

                channel.port1.postMessage({ chunk: value }, [value.buffer]);
            }
        })();
    }
    
    return handler;
}

form.addEventListener("submit", async e => {
    e.preventDefault();
    const sw = await navigator.serviceWorker.ready;

    await window.sodium.ready;

    const formData = new FormData(form);
    const userId = formData.get("user-id").toString();
    const fileId = formData.get("file-id").toString();

    const [key, header, nonce] = window.location.hash.substring(1).split(".").map(v => fromBase64Url(v));

    console.log("Key", key)
    console.log("Header", header);
    console.log("Nonce", nonce);

    // eslint-disable-next-line no-undef
    const manager = new ChunkedDownloadManager(key, header, nonce, userId, fileId);
    const totalFileSize = await manager.getTotalSize();

    let worker;
    worker = await createServiceWorkerHandler(sw.active, fileId);
    
    manager.start(worker);

    while (true) {
        const progress = manager.progress / totalFileSize;
        setLoader(progress);
        await new Promise(resolve => setTimeout(resolve, 100));

        if (progress >= 1) {
            break;
        }
    }
});
