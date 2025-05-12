
/**
 * 
 * @param {ServiceWorker} sw
 * @param {string} fileId
 */
const createServiceWorkerHandler = async (sw, fileId) => {
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
                    iframe.src = `../../dl/${fileId}`;
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

                if (!value) {
                    continue;
                }

                channel.port1.postMessage({ chunk: value }, [value.buffer]);
            }
        })();
    }
    
    return handler;
}
