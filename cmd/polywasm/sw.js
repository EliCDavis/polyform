// https://developer.mozilla.org/en-US/docs/Web/API/Service_Worker_API/Using_Service_Workers

// Note the 'go.1.23.4' below, that matches the version you just found:
importScripts('https://cdn.jsdelivr.net/gh/golang/go@go1.23.4/misc/wasm/wasm_exec.js')
// If you compiled with TinyGo then, similarly, use:
// importScripts('https://cdn.jsdelivr.net/gh/tinygo-org/tinygo@0.35.0/targets/wasm_exec.js')

importScripts('https://cdn.jsdelivr.net/gh/nlepage/go-wasm-http-server@v2.1.0/sw.js')


const CURRENT_VERSION = 'v0.0.23';
const WASM = 'main.wasm.gz'

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

registerWasmHTTPListener(WASM)