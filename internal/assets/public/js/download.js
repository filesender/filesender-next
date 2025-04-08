const form = document.querySelector("form");

navigator.serviceWorker.register("../../sw.js").then(async e => {
    console.log('Service Worker registered!');

    navigator.serviceWorker.addEventListener("message", e => {
        if (e.data.url) {
            console.log('SW Controller:', navigator.serviceWorker.controller);
            console.log(e.data.url);

            const a = document.createElement("a");
            a.href = e.data.url;
            a.download = "";
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
        }
    });
});



form.addEventListener("submit", async e => {
    e.preventDefault();

    const sw = await navigator.serviceWorker.ready;
    const untar = new Untar(sw);
    untar.start();
    
    const response = await fetch(e.target.action);
    const reader = response.body.getReader();

    while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        untar.push(value);
    }

    untar.end();
});
