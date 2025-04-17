
class ChunkedDownloadManager {
    /**
     * 
     * @param {ServiceWorker} sw 
     * @param {Uint8Array} key 
     * @param {Uint8Array} header 
     */
    constructor(sw, key, header) {
        this.sw = sw;
        this.key = key;
        this.header = header;
    }

    
}
