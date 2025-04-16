
const form = document.querySelector("form");

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

console.log(form);

form.addEventListener("submit", async e => {
    e.preventDefault();
    await window.sodium.ready;
    const sodium = window.sodium;

    const formData = new FormData(form);
    const userId = formData.get("user-id");
    const fileId = formData.get("file-id");

    const [key, header] = window.location.hash.substring(1).split(".").map(v => fromBase64Url(v));

    console.log(userId, fileId);
    console.log(key, header);
});
