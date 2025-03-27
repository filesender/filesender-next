const form = document.querySelector("form");
var transferId = null;

/**
 * Dummy error handling function
 * @param {string} msg Message to show to use
 */
const showError = msg => {
    console.log(`ERROR: ${msg}`);
}

/**
 * Uploads a file to a transfer
 * @param {string} expiryDate `YYYY-MM-DD` formatted expiry date of the transfer
 * @param {File} file 
 * @returns {Promise<bool>} If the transfer was successful or not
 */
const uploadFile = async (expiryDate, file) => {
    const formData = new FormData();
    formData.append("file", file);
    formData.append("expiry-date", expiryDate);

    if (transferId !== null) {
        formData.append("transfer-id", transferId);
    }
    
    const response = await fetch("api/v1/upload", {
        method: "POST",
        body: formData
    });

    if (response.status === 200) {
        transferId = response.url.split('upload/')[1];
        return true
    }

    showError("Something went wrong uploading file");
    console.error(response.body)
    return false;
}

form.addEventListener("submit", async e => {
    e.preventDefault();

    const formData = new FormData(form);
    const filesInput = formData.getAll("file");
    const expiryDate = formData.get("expiry-date");

    if (filesInput[0].size === 0) {
        return showError("You have to select files");
    }

    var files = [
        ...filesInput.filter(f => f.size !== 0)
    ];

    for (let i = 0; i < files.length; i++) {
        let tries = 0;
        while (tries < 3) {
            try {
                await uploadFile(expiryDate, files[i]);
                break;
            } catch (e) {
                console.error("Error uploading", e);
                tries++;
            }
        }
    }

    window.location.replace(`upload/${transferId}`);
});
