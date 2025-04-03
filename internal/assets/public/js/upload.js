const form = document.querySelector("form");
var userId = "";

document.body.onload = () => {
    const filesSelector = document.querySelector("#files-selector");
    const label = filesSelector.parentElement.querySelector("label");

    filesSelector.setAttribute("multiple", "true");
    label.innerText = "Select files";
}

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

/**
 * Archives files and returns a .tar Blob object
 * This is a placeholder function for now..
 * @param {FileList | File[]} files - The files to be added to the tar archive
 * @returns {Promise<Blob>} A blob containing the FULL .tar (not chunked or a stream)
 */
const archiveFiles = async (files) => {
    const stream = new ReadableStream({
        async start(controller) {
            const generator = generateTarStream(files);
            for await (const chunk of generator) {
                console.log(chunk);
                controller.enqueue(chunk);
            }
            controller.close();
        }
    });

    const response = new Response(stream, {
        headers: {
            "Content-Type": "application/x-tar"
        }
    });
    return response.blob();
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

    const tarBlob = await archiveFiles(files);
    const file = new File([tarBlob], "archive.tar");
    
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
