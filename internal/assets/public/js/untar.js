
class Untar {
    constructor(sw) {
        this.sw = sw;
        this.buffer = new Uint8Array(0);
        this.resolveNext = null;
        this.ended = false;
    }

    /**
     * 
     * @param {Uint8Array} chunk 
     */
    push(chunk) {
        const newBuffer = new Uint8Array(this.buffer.length + chunk.length);
        newBuffer.set(this.buffer);
        newBuffer.set(chunk, this.buffer.length);
        this.buffer = newBuffer;

        if (this.resolveNext) {
            const resolve = this.resolveNext;
            this.resolveNext = null;
            resolve();
        }
    }

    end() {
        this.ended = true;

        if (this.resolveNext) {
            const resolve = this.resolveNext;
            this.resolveNext = null;
            resolve();
        }
    }

    /**
     * 
     * @param {number} len 
     */
    async waitFor(len) {
        while (this.buffer.length < len) {
            if (this.ended) break;
            await new Promise(res => this.resolveNext = res);
        }
        return this.buffer.length >= len;
    }

    async start() {
        const TAR_BLOCK_SIZE = 512;

        /**
         * 
         * @param {Uint8Array} buf 
         * @param {number} start 
         * @param {number} len 
         * @returns 
         */
        const getString = (buf, start, len) =>
            new TextDecoder().decode(buf.slice(start, start + len)).replace(/\0+$/, '');

        /**
         * 
         * @param {Uint8Array} buf 
         * @param {number} start 
         * @param {number} len 
         * @returns 
         */
        const getOctal = (buf, start, len) =>
            parseInt(getString(buf, start, len).trim() || '0', 8);

        while (true) {
            if (!(await this.waitFor(TAR_BLOCK_SIZE))) break;

            const header = this.buffer.slice(0, TAR_BLOCK_SIZE);
            const name = getString(header, 0, 100);
            const size = getOctal(header, 124, 12);
            const id = crypto.randomUUID();

            if (!name) break;

            this.buffer = this.buffer.slice(TAR_BLOCK_SIZE);

            const { readable, writable } = new TransformStream();
            const writer = writable.getWriter();

            this.sw.active.postMessage({ id, name, readableStream: readable }, [readable]);            

            let bytesLeft = size;
            while (bytesLeft > 0) {
                const toWrite = Math.min(bytesLeft, this.buffer.length);
                const chunk = this.buffer.slice(0, toWrite);
                await writer.write(chunk);
                this.buffer = this.buffer.slice(toWrite);
                bytesLeft -= toWrite;

                if (bytesLeft > 0 && this.buffer.length === 0) {
                    if (!(await this.waitFor(1))) break;
                }
            }

            await writer.close();

            const padding = (TAR_BLOCK_SIZE - (size % TAR_BLOCK_SIZE)) % TAR_BLOCK_SIZE;
            if (padding > 0) {
                if (!(await this.waitFor(padding))) break;
                this.buffer = this.buffer.slice(padding);
            }
        }
    }
}
