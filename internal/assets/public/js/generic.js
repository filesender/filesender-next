/* eslint-disable no-unused-vars */
const progressHolder = document.querySelector("div#progress");
const errorBox = document.querySelector("div.error");
const progress = document.querySelector("progress");
const loaderText = document.querySelector("p#progress");

/**
 * Convert a byte count to a human readable string
 *
 * @param {number} bytes
 * @returns {string}
 */
function formatBytes(bytes) {
    if (!Number.isFinite(bytes) || bytes < 0) {
        throw new Error("Input incorrect");
    }

    const UNITS = ["bytes", "KB", "MB", "GB", "TB"];
    const BASE = 1024;

    let unitIndex = 0;
    let value = bytes;

    while (value >= BASE && unitIndex < UNITS.length - 1) {
        value /= BASE;
        unitIndex++;
    }

    return `${value.toFixed(0)} ${UNITS[unitIndex]}`;
}

/**
 * Error showing
 * @param {string} msg Message to show to use
 */
const showError = msg => {
    console.log(`ERROR: ${msg}`);
    errorBox.innerText = msg;
    errorBox.classList.remove("hidden");
}

/**
 * Hides whatever current error message is being shown
 */
const hideError = () => {
    errorBox.classList.add("hidden");
}

/**
 * 
 * @param {number} current 
 * @param {number} max 
 */
const setProgress = (current, max) => {
    if (current === 0) return;

    progressHolder.classList.remove("hidden");
    progress.max = max;
    progress.value = current;
    
    const currentText = formatBytes(current);
    const maxText = formatBytes(max);
    loaderText.innerText = `${currentText} / ${maxText}`;
}
