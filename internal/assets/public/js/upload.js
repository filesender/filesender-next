const form = document.querySelector("form");

/**
 * Dummy error handling function
 * @param {string} msg Message to show to use
 */
const showError = msg => {
    console.log(`ERROR: ${msg}`);
}

/**
 * Sends a request to initialise a new transfer before uploading files
 * @param {{subject: string|null, message: string|null, expiry_date: string|null}} requestData Transfer initialisation data
 * @returns {Promise<false | {id: number, user_id: string, file_count: number, total_byte_size: number, subject: string, message: string, download_count: number, expiry_date: string, creation_date: string}>} Either returns false when request failed, or the newly created transfer object
 */
const createTransfer = async (requestData) => {
    const response = await fetch("api/v1/transfer", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(requestData)
    });
    const data = await response.json();
    
    if (data.success) {
        return data.data.transfer;
    }
    return false;
}

/**
 * Uploads a file to a transfer
 * @param {number} transferId The transfer ID which the file is being uploaded to
 * @param {File} file 
 * @returns {Promise<bool>} If the transfer was successful or not
 */
const uploadFile = async (transferId, file) => {
    const formData = new FormData();
    formData.append("transfer_id", transferId);
    formData.append("file", file);

    if (file.webkitRelativePath) {
        if (file.webkitRelativePath.split("/").slice(1, -1).join("/") !== "") {
            showError("Not uploading relative files");
            return false;
        }
    }
    
    const response = await fetch("api/v1/upload", {
        method: "POST",
        body: formData
    });
    const data = await response.json();

    if (data.success) {
        return true
    }

    showError(data.message);
    return false;
}

form.addEventListener("submit", async e => {
    e.preventDefault();

    const formData = new FormData(form);
    const folderInput = formData.getAll("folder-input");
    const filesInput = formData.getAll("files-input");
    const subject = formData.get("subject");
    const message = formData.get("message");
    const expiryDate = formData.get("expiry-date") + "T00:00:00Z";

    if (folderInput[0].size === 0 && filesInput[0].size === 0) {
        return showError("You have to select files");
    }

    var files = [
        ...folderInput.filter(f => f.size !== 0),
        ...filesInput.filter(f => f.size !== 0)
    ];

    const transfer = await createTransfer({
        subject: subject === "" ? null : subject,
        message: message === "" ? null : message,
        "expiry_date": expiryDate
    });

    for (let i = 0; i < files.length; i++) {
        let tries = 0;
        while (tries < 3) {
            try {
                await uploadFile(transfer.id, files[i]);
                break;
            } catch (e) {
                console.error("Error uploading", e);
                tries++;
            }
        }
    }

    window.location.replace(`upload/${transfer.id}`);
});
