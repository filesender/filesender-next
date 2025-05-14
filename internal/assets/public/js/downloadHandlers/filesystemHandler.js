
/**
 * 
 * @param {() => void} ready
 */
// eslint-disable-next-line no-unused-vars
const createFileSystemHandler = (ready) => {
        /**
     * 
     * @param {string} fileName 
     * @param {ReadableStream} stream 
     */
    const handler = (fileName, stream) => {
        (async () => {
            const handle = await window.showSaveFilePicker({
                startIn: "downloads",
                suggestedName: fileName
            });
            if (!handle) return;

            const writer = await handle.createWritable();
            stream.pipeTo(writer);

            ready();
        })();
    }

    return handler;
}
