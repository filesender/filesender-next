
/**
 * 
 * @param {() => void} ready
 */
// eslint-disable-next-line no-unused-vars
const createMemoryHandler = (ready) => {
    /**
     * @param {DownloadManager} manager
     * @param {string} fileName 
     * @param {ReadableStream} stream 
     */
    const handler = (_manager, fileName, stream) => {
        ready();

        const reader = stream.getReader();
        (async () => {
            const chunks = [];

            while (true) {
                const { done, value } = await reader.read();
                if (done) break;
                chunks.push(value);
            }

            const blob = new Blob(chunks);
            const url = URL.createObjectURL(blob);

            const link = document.createElement("a");
            link.href = url;
            link.download = fileName;
            link.click();

            URL.revokeObjectURL(url);
        })();
    }

    return handler;
}
