
const form = document.querySelector("form");

// Register service worker
navigator.serviceWorker.register("../../sw.js").then(async e => {
    console.log('Service Worker registered!');

    navigator.serviceWorker.addEventListener("message", e => {
        if (e.data.url) {
            console.log('SW Controller:', navigator.serviceWorker.controller);
            console.log(e.data.url);

            const a = document.createElement("a");
            a.href = e.data.url;
            a.download = "";
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
        }
    });
});

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

const getFileInfo = async (userId, fileId) => {
    const response = await fetch(`../../api/v1/download/${userId}/${fileId}`, {
        method: "HEAD"
    });
    
    return {
        available: response.headers.get("available") === "true",
        chunked: response.headers.get("chunked") === "true",
        chunkCount: parseInt(response.headers.get("chunk-count"))
    }
}

form.addEventListener("submit", async e => {
    e.preventDefault();
    const sw = await navigator.serviceWorker.ready;

    await window.sodium.ready;
    const sodium = window.sodium;

    const formData = new FormData(form);
    const userId = formData.get("user-id").toString();
    const fileId = formData.get("file-id").toString();

    const fileInfo = await getFileInfo(userId, fileId);
    sw.active.postMessage({
        type: "info",
        info: {
            userId,
            fileId,
            ...fileInfo
        }
    });

    const [key, header] = window.location.hash.substring(1).split(".").map(v => fromBase64Url(v));

    console.log(userId, fileId);
    console.log(key, header);
});
