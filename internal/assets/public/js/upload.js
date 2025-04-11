const form = document.querySelector("form");
var userId = "";

/**
 * Dummy error handling function
 * @param {string} msg Message to show to use
 */
const showError = msg => {
    console.log(`ERROR: ${msg}`);
}

/**
 * Uploads a file
 * @param {string} expiryDate `YYYY-MM-DD` formatted expiry date of the file
 * @param {File} file 
 * @returns {Promise<string|false>} Contains file ID with successful, otherwise `false`
 */
const uploadFile = async (expiryDate, file) => {
    const formData = new FormData();
    formData.append("file", file);
    formData.append("expiry-date", expiryDate);
    
    const response = await fetch("api/v1/upload", {
        method: "POST",
        body: formData
    });

    if (response.status === 200) {
        const parials = response.url.split('download/')[1];
        return parials.split("/")[1];
    }

    showError("Something went wrong uploading file");
    console.error(response.body)
    return false;
}

form.addEventListener("submit", async e => {
    e.preventDefault();

    const formData = new FormData(form);
    const file = formData.get("file");
    const expiryDate = formData.get("expiry-date");

    if (file.name === "") {
        return showError("You have to select a file");
    }
    
    let tries = 0;
    var fileId;
    while (tries < 3) {
        try {
            fileId = await uploadFile(expiryDate, file);
            break;
        } catch (e) {
            console.error("Error uploading", e);
            tries++;
        }
    }

    if (fileId !== false)
        window.location.replace(`download/${userId}/${fileId}`);
});
