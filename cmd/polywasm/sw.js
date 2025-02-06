// https://developer.mozilla.org/en-US/docs/Web/API/Service_Worker_API/Using_Service_Workers

// Note the 'go.1.23.4' below, that matches the version you just found:
importScripts('https://cdn.jsdelivr.net/gh/golang/go@go1.23.4/misc/wasm/wasm_exec.js')

// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
// The base of this code comes from:
// 
// https://github.com/nlepage/go-wasm-http-server
//
// Apache License
// Version 2.0, January 2004
// http://www.apache.org/licenses/
//
// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
function registerWasmHTTPListener(wasm, { base, passthroughFunc, cacheName, args = [] } = {}) {
    let path = new URL(registration.scope).pathname
    if (base && base !== '') path = `${trimEnd(path, '/')}/${trimStart(base, '/')}`

    const handlerPromise = new Promise(setHandler => {
        self.wasmhttp = {
            path,
            setHandler,
        }
    })

    const go = new Go()
    go.argv = [wasm, ...args]
    const source = cacheName
        ? caches.open(cacheName).then((cache) => cache.match(wasm)).then((response) => response ?? fetch(wasm))
        : caches.match(wasm).then(response => (response) ?? fetch(wasm))
    WebAssembly.instantiateStreaming(source, go.importObject).then(({ instance }) => go.run(instance))

    addEventListener('fetch', e => {
        const { pathname } = new URL(e.request.url);

        if (passthroughFunc && passthroughFunc(e.request)) {
            e.respondWith(fetch(e.request))
            return;
        }

        if (!pathname.startsWith(path)) {
            // e.respondWith(fetch(e.request))
            return;
        }

        e.respondWith(handlerPromise.then(handler => handler(e.request)))
    })
}

function trimStart(s, c) {
    let r = s
    while (r.startsWith(c)) r = r.slice(c.length)
    return r
}

function trimEnd(s, c) {
    let r = s
    while (r.endsWith(c)) r = r.slice(0, -c.length)
    return r
}
// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

const CURRENT_VERSION = "{{.Version}}";
const WASM = 'main.wasm'

addEventListener("install", (event) => {
    console.log("installed", event)
    event.waitUntil(caches.open(CURRENT_VERSION).then((cache) => cache.add(WASM)))
});

const deleteCache = async (key) => {
    await caches.delete(key);
};

const deleteOldCaches = async () => {
    const cacheKeepList = [CURRENT_VERSION];
    const keyList = await caches.keys();
    const cachesToDelete = keyList.filter((key) => !cacheKeepList.includes(key));
    await Promise.all(cachesToDelete.map(deleteCache));
};

addEventListener('activate', event => {
    console.log("activate", event)
    event.waitUntil(deleteOldCaches());
    event.waitUntil(clients.claim())
})

addEventListener('message', (event) => {
    console.log(`The service worker sent me a message: ${event.data}`);
})

const approvedFiles = [
    "index.html",
    "sw.js"
]

function isSameOrigin(urlString) {
    const urlOrigin = (new URL(urlString)).origin;
    return urlOrigin === self.location.origin;
}

function urlIsRootServiceWorkerInstall(urlString) {
    // 
    let cleanString = urlString;
    if (!cleanString.endsWith("/")) {
        cleanString += "/"
    }

    return self.location.href.slice(0, -5) === cleanString
}

registerWasmHTTPListener(
    WASM,
    {
        passthroughFunc: (request) => {
            let url = new URL(request.url);

            if (urlIsRootServiceWorkerInstall(request.url)) {
                return true;
            }

            for (let i = 0; i < approvedFiles.length; i++) {
                if (url.pathname.endsWith(approvedFiles[i])) {
                    return true;
                }
            }

            return !isSameOrigin(request.url);
        }
    }
)