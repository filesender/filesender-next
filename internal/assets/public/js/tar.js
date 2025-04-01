
/**
 * Pads a string to the specified length using the given padding character.
 * If the string is longer than the desired length, it is shortened.
 * 
 * @param {string} str - The string to pad.
 * @param {number} length - The desired total length of the output string.
 * @param {string} [padding=' '] - The character to use for padding.
 * @returns {string} - The padded string.
 */
const pad = (str, length, padding = ' ') => {
    str = String(str);
    if (str.length > length) return str.slice(0, length);
    return str + padding.repeat(length - str.length);
};

/**
 * Converts a number to an octal string and pads it to the given length,
 * ending with a null character as required by the TAR format.
 * 
 * @param {number} number - The number to convert to octal.
 * @param {number} length - The total length of the resulting string.
 * @returns {string} - The padded octal string ending in '\0'.
 */
const padOctal = (number, length) => {
    return number.toString(8).padStart(length - 1, '0') + '\0';
};

/**
 * Asynchronously generates a TAR archive stream from an array of File objects.
 * This is a generator that yields Uint8Array chunks as the archive is built.
 * 
 * @param {FileList | File[]} files - The array or FileList of File objects to include in the TAR archive.
 * @yields {Uint8Array} - Chunks of the TAR file as it's being generated.
 */
const generateTarStream = async function* (files) {
    const encoder = new TextEncoder();

    for (const file of files) {
        const name = pad(file.name, 100, '\0');
        const mode = padOctal(0o644, 8);
        const uid = padOctal(0, 8);
        const gid = padOctal(0, 8);
        const size = padOctal(file.size, 12);
        const modified = padOctal(Math.floor(file.lastModified / 1000), 12);
        const checksumPlaceholder = '        ';
        const typeflag = '0';
        const linkname = pad('', 100, '\0');
        const magic = 'ustar\0';
        const version = '00';
        const uname = pad('user', 32, '\0');
        const gname = pad('group', 32, '\0');
        const devmajor = padOctal(0, 8);
        const devminor = padOctal(0, 8);
        const prefix = pad('', 155, '\0');

        // Combine header parts
        const headerStr = name + mode + uid + gid + size + modified + checksumPlaceholder + typeflag + linkname + magic + version + uname + gname + devmajor + devminor + prefix;

        // Make sure it's 512 bytes
        let headerBytes = encoder.encode(headerStr);
        if (headerBytes.length < 512) {
            const padded = new Uint8Array(512);
            padded.set(headerBytes);
            headerBytes = padded;
        }

        // Calculate checksum (all bytes with checksum field set to space)
        for (let i = 148; i < 156; i++) headerBytes[i] = 32; // space

        const checksum = padOctal(
            headerBytes.reduce((sum, byte) => sum + byte, 0),
            8
        );

        const checksumBytes = encoder.encode(checksum);
        headerBytes.set(checksumBytes, 148); // inject real checksum

        yield headerBytes;

        // Stream file contents
        const reader = file.stream().getReader();
        let total = 0;
        while (true) {
            const { done, value } = await reader.read();
            if (done) break;
            total += value.length;
            yield value;
        }

        // Pad to 512-byte block
        const remainder = total % 512;
        if (remainder !== 0) {
            yield new Uint8Array(512 - remainder);
        }
    }

    // Tar file has to end with 2 512 byte blocks
    yield new Uint8Array(512);
    yield new Uint8Array(512);
}
