
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


class RequestManager {
    constructor() {
    }

    fetchText(theUrl, callback) {
        const xmlHttp = new XMLHttpRequest();
        xmlHttp.onreadystatechange = () => {
            if (xmlHttp.readyState == 4 && xmlHttp.status == 200) {
                callback(xmlHttp.responseText);
            }
        }
        xmlHttp.open("GET", theUrl, true); // true for asynchronous 
        xmlHttp.send(null);
    }

    fetch(theUrl, callback) {
        const xmlHttp = new XMLHttpRequest();
        xmlHttp.onreadystatechange = () => {
            if (xmlHttp.readyState == 4 && xmlHttp.status == 200) {
                console.log(xmlHttp.responseText)
                callback(JSON.parse(xmlHttp.responseText));
            }
        }
        xmlHttp.open("GET", theUrl, true); // true for asynchronous 
        xmlHttp.send(null);
    }

    post(theUrl, body, callback) {
        const xmlHttp = new XMLHttpRequest();
        xmlHttp.onreadystatechange = function () {
            if (xmlHttp.readyState == 4 && xmlHttp.status == 200) {
                callback(JSON.parse(xmlHttp.responseText));
            }
        }
        xmlHttp.open("POST", theUrl, true); // true for asynchronous 
        xmlHttp.send(JSON.stringify(body));
    }

    getStartedTime(callback) {
        this.fetch("/started", callback);
    }

    getSchema(callback) {
        this.fetch("/schema", callback);
    }

    updateProfile(data, callback) {
        this.post("/profile", data, callback);
    }
}