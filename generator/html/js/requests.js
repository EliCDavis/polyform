
function download(theUrl, callback) {
    const xmlHttp = new XMLHttpRequest();
    xmlHttp.responseType = 'blob';
    xmlHttp.onreadystatechange = function () {
        if (xmlHttp.readyState == 4 && xmlHttp.status == 200) {
            console.log(xmlHttp)
            callback(xmlHttp.response);
        }
    }
    xmlHttp.open("GET", theUrl, true); // true for asynchronous 
    xmlHttp.send(null);
}

function saveFileToDisk(theUrl, fileName) {
    download(theUrl, (data) => {
        const a = document.createElement('a');
        a.download = fileName;
        const url = window.URL.createObjectURL(data);
        a.href = url;
        a.click();
        window.URL.revokeObjectURL(url);
    })
}

class RequestManager {
    constructor() {
    }

    fetchText(theUrl, successCallback, errorCallback) {
        const xmlHttp = new XMLHttpRequest();
        xmlHttp.onreadystatechange = () => {
            if (xmlHttp.readyState == 4) {
                if (xmlHttp.status == 200) {
                    successCallback(xmlHttp.responseText);
                } else if (errorCallback) {
                    errorCallback(xmlHttp.responseText);
                }
            }
        }
        xmlHttp.open("GET", theUrl, true); // true for asynchronous 
        xmlHttp.send(null);
    }

    fetch(theUrl, callback) {
        const xmlHttp = new XMLHttpRequest();
        xmlHttp.onreadystatechange = () => {
            if (xmlHttp.readyState == 4 && xmlHttp.status == 200) {
                // console.log(xmlHttp.responseText)
                callback(JSON.parse(xmlHttp.responseText));
            }
        }
        xmlHttp.open("GET", theUrl, true); // true for asynchronous 
        xmlHttp.send(null);
    }

    post(theUrl, body, callback) {
        const xmlHttp = new XMLHttpRequest();
        xmlHttp.onreadystatechange = function () {
            if (xmlHttp.readyState == 4 && xmlHttp.status == 200 && callback) {
                callback(JSON.parse(xmlHttp.responseText));
            }
        }
        xmlHttp.open("POST", theUrl, true); // true for asynchronous 
        xmlHttp.send(JSON.stringify(body));
    }

    postBinary(theUrl, body, callback) {
        const xmlHttp = new XMLHttpRequest();
        xmlHttp.onreadystatechange = function () {
            if (xmlHttp.readyState == 4 && xmlHttp.status == 200 && callback) {
                callback(JSON.parse(xmlHttp.responseText));
            }
        }
        xmlHttp.open("POST", theUrl, true); // true for asynchronous 
        xmlHttp.send(body);
    }

    getStartedTime(callback) {
        this.fetch("/started", callback);
    }

    getSchema(callback) {
        this.fetch("/schema", callback);
    }

    updateProfile(key, data, binary, callback) {
        if (binary) {
            this.postBinary("/profile/" + key, data, callback);
        } else {
            this.post("/profile/" + key, data, callback);
        }
    }
}