
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

    fetchImage(imgUrl, successCallback, errorCallback) {
        const img = document.createElement("img");
        // ? + perfoance.now() is to break browswer cache
        img.src = imgUrl + "?" + performance.now();
        img.onload = () => {
            successCallback(img);
        };
        img.onerror = (event) => {
            errorCallback(event)
        }
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


    fetchJSON(theUrl, callback) {
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

    fetchRaw(theUrl, callback) {
        const xmlHttp = new XMLHttpRequest();
        xmlHttp.responseType = 'blob';

        xmlHttp.onreadystatechange = () => {
            if (xmlHttp.readyState == 4 && xmlHttp.status == 200) {
                // console.log(xmlHttp.responseText)
                callback(xmlHttp.response);
            }
        }
        xmlHttp.open("GET", theUrl, true); // true for asynchronous 
        xmlHttp.send(null);
    }

    deleteEmptyBodyEmptyResponse(theUrl, callback) {
        const xmlHttp = new XMLHttpRequest();
        // xmlHttp.responseType = 'blob';

        xmlHttp.onreadystatechange = () => {
            if (xmlHttp.readyState == 4 && xmlHttp.status == 200) {
                // console.log(xmlHttp.responseText)
                // callback(xmlHttp.response);
                callback();
            }
        }
        xmlHttp.open("DELETE", theUrl, true); // true for asynchronous 
        xmlHttp.send(null);
    }

    postJsonBodyJsonResponse(theUrl, body, callback) {
        this.postBinaryJsonResponse(theUrl, JSON.stringify(body), callback)
    }

    postJsonBodyEmptyResponse(theUrl, body, callback) {
        const xmlHttp = new XMLHttpRequest();
        xmlHttp.onreadystatechange = () => {
            if (xmlHttp.readyState == 4 && xmlHttp.status == 200 && callback) {
                callback();
            }
        }
        xmlHttp.open("POST", theUrl, true); // true for asynchronous 
        xmlHttp.send(JSON.stringify(body));
    }

    postTextBodyEmptyResponse(theUrl, body, callback) {
        const xmlHttp = new XMLHttpRequest();
        xmlHttp.onreadystatechange = () => {
            if (xmlHttp.readyState == 4 && xmlHttp.status == 200 && callback) {
                callback();
            }
        }
        xmlHttp.open("POST", theUrl, true); // true for asynchronous 
        xmlHttp.send(body);
    }

    postBinaryEmptyResponse(theUrl, body, callback) {
        const xmlHttp = new XMLHttpRequest();
        xmlHttp.onreadystatechange = () => {
            if (xmlHttp.readyState == 4 && xmlHttp.status == 200 && callback) {
                callback();
            }
        }
        xmlHttp.open("POST", theUrl, true); // true for asynchronous 
        xmlHttp.send(body);
    }

    postBinaryJsonResponse(theUrl, body, callback) {
        const xmlHttp = new XMLHttpRequest();
        xmlHttp.onreadystatechange = () => {
            if (xmlHttp.readyState == 4 && xmlHttp.status == 200 && callback) {
                callback(JSON.parse(xmlHttp.responseText));
            }
        }
        xmlHttp.open("POST", theUrl, true); // true for asynchronous 
        xmlHttp.send(body);
    }

    deleteJSONBodyJSONResponse(theUrl, body, callback) {
        const xmlHttp = new XMLHttpRequest();
        xmlHttp.onreadystatechange = () => {
            if (xmlHttp.readyState == 4 && xmlHttp.status == 200 && callback) {
                callback(JSON.parse(xmlHttp.responseText));
            }
        }
        xmlHttp.open("DELETE", theUrl, true); // true for asynchronous 
        xmlHttp.send(JSON.stringify(body));
    }

    getStartedTime(callback) {
        this.fetchJSON("./started", callback);
    }

    getSchema(callback) {
        this.fetchJSON("./schema", callback);
    }

    setParameter(key, data, binary, callback) {
        const url = "./parameter/value/" + key;
        if (binary) {
            this.postBinaryEmptyResponse(url, data, callback);
        } else {
            this.postJsonBodyEmptyResponse(url, data, callback);
        }
    }

    setParameterTitle(inNodeID, value, callback) {
        this.postTextBodyEmptyResponse("./parameter/name/" + inNodeID, value, callback);
    }

    setParameterInfo(inNodeID, value, callback) {
        this.postTextBodyEmptyResponse("./parameter/description/" + inNodeID, value, callback);
    }

    setProducerTitle(inNodeID, value, callback) {
        this.postTextBodyEmptyResponse("./producer/name/" + inNodeID, value, callback);
    }

    getParameterValue(key, callback) {
        this.fetchRaw("./parameter/value/" + key, callback);
    }

    deleteNodeInput(nodeID, inputPortName, callback) {
        this.deleteJSONBodyJSONResponse("node/connection", {
            "nodeId": nodeID,
            "inPortName": inputPortName
        }, callback)
    }

    setNodeInputConnection(inNodeID, inputPortName, outNodeID, outPortName, callback) {
        this.postJsonBodyJsonResponse("node/connection", {
            "nodeOutId": outNodeID,
            "outPortName": outPortName,
            "nodeInId": inNodeID,
            "inPortName": inputPortName
        }, callback)
    }

    setNodeMetadata(inNodeID, key, metadata, callback) {
        this.postJsonBodyJsonResponse(`graph/metadata/nodes/${inNodeID}/${key}`, metadata, callback)
    }



    createNote(noteID, note, callback) {
        this.postJsonBodyJsonResponse(`graph/metadata/notes/${noteID}`, note, callback)
    }

    setNoteMetadata(noteID, key, metadata, callback) {
        this.postJsonBodyJsonResponse(`graph/metadata/notes/${noteID}/${key}`, metadata, callback)
    }

    deleteMetadata(path, callback) {
        this.deleteEmptyBodyEmptyResponse(`graph/metadata/${path}`, callback)
    }

    createNode(nodeType, callback) {
        this.postJsonBodyJsonResponse("node", {
            "nodeType": nodeType,
        }, callback)
    }

    deleteNode(nodeId, callback) {
        this.deleteJSONBodyJSONResponse("node", {
            "nodeID": nodeId,
        }, callback)
    }

    getGraph(callback) {
        this.fetchJSON("./graph", callback)
    }

    getSwagger(callback) {
        this.fetchJSON("./swagger", callback)
    }

    setGraph(newGraph, callback) {
        this.postJsonBodyJsonResponse("./graph", newGraph, callback)
    }
}