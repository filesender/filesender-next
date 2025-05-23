
/**
 * 
 * @param {() => void} ready
 * @param {string} fileId
 * @param {ServiceWorker} sw
 */
// eslint-disable-next-line no-unused-vars
const createServiceWorkerHandler = (ready, fileId, sw) => {
    /**
     * @param {DownloadManager} manager
     * @param {string} fileName 
     * @param {ReadableStream} stream 
     */
    const handler = (manager, fileName, stream) => {
        sw.postMessage({
            type: "delete",
            id: fileId
        });

        const broadcast = new BroadcastChannel(fileId);
        broadcast.addEventListener("message", e => {
            if (e.data.type === "downloadAvailable") {
                if (e.data.id === fileId) {
                    const iframe = document.createElement('iframe');
                    iframe.style.display = 'none';
                    iframe.src = `../../dl/${fileId}`;
                    document.body.appendChild(iframe);

                    ready();
                }
            }

            if (e.data.type === "downloadCancelled") {
                if (!manager.cancelled) {
                    manager.cancel();
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

                if (!value) {
                    continue;
                }

                channel.port1.postMessage({ chunk: value }, [value.buffer]);
            }
        })();
    }
    
    return handler;
}
