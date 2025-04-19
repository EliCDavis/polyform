import { GraphInstance, Manifest, NodeInstance, NodeType } from "./schema";

export function downloadBlob(theUrl: string, callback: (body: any) => void): void {
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

export function saveFileToDisk(theUrl: string, fileName: string): void {
    downloadBlob(theUrl, (data) => {
        const a = document.createElement('a');
        a.download = fileName;
        const url = window.URL.createObjectURL(data);
        a.href = url;
        a.click();
        window.URL.revokeObjectURL(url);
    })
}

export interface SetProducerBody {
    nodePort: string,
    producer: string,
}

export interface StartedResponse {
    time: string,
    modelVersion: number
}

export interface CreateNodeResponse {
    nodeID: string;
    data: NodeInstance;
}

type ResponseCallback<T> = (responseBody: T) => void

export class RequestManager {
    constructor() {
    }

    fetchImage(imgUrl: string, successCallback, errorCallback): void {
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

    fetchText(theUrl: string, successCallback, errorCallback?: (data: string) => void): void {
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


    fetchJSON(theUrl: string, callback): void {
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

    fetchRaw(theUrl: string, callback): void {
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

    deleteEmptyBodyEmptyResponse(theUrl: string, callback): void {
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

    postJsonBodyJsonResponse(theUrl: string, body: any, callback?: ResponseCallback<any>): void {
        this.postBinaryJsonResponse(theUrl, JSON.stringify(body), callback)
    }

    postJsonBodyEmptyResponse(theUrl: string, body, callback): void {
        const xmlHttp = new XMLHttpRequest();
        xmlHttp.onreadystatechange = () => {
            if (xmlHttp.readyState == 4 && xmlHttp.status == 200 && callback) {
                callback();
            }
        }
        xmlHttp.open("POST", theUrl, true); // true for asynchronous 
        xmlHttp.send(JSON.stringify(body));
    }

    postTextBodyEmptyResponse(theUrl: string, body, callback?: () => void): void {
        const xmlHttp = new XMLHttpRequest();
        xmlHttp.onreadystatechange = () => {
            if (xmlHttp.readyState == 4 && xmlHttp.status == 200 && callback) {
                callback();
            }
        }
        xmlHttp.open("POST", theUrl, true); // true for asynchronous 
        xmlHttp.send(body);
    }

    postBinaryEmptyResponse(theUrl: string, body, callback): void {
        const xmlHttp = new XMLHttpRequest();
        xmlHttp.onreadystatechange = () => {
            if (xmlHttp.readyState == 4 && xmlHttp.status == 200 && callback) {
                callback();
            }
        }
        xmlHttp.open("POST", theUrl, true); // true for asynchronous 
        xmlHttp.send(body);
    }

    postBinaryJsonResponse(theUrl: string, body: any, callback?: ResponseCallback<any>): void {
        const xmlHttp = new XMLHttpRequest();
        xmlHttp.onreadystatechange = () => {
            if (xmlHttp.readyState == 4 && xmlHttp.status == 200 && callback) {
                callback(JSON.parse(xmlHttp.responseText));
            }
        }
        xmlHttp.open("POST", theUrl, true); // true for asynchronous 
        xmlHttp.send(body);
    }

    deleteJSONBodyJSONResponse(theUrl: string, requestBody: any, callback?: ResponseCallback<any>): void {
        const xmlHttp = new XMLHttpRequest();
        xmlHttp.onreadystatechange = () => {
            if (xmlHttp.readyState == 4 && xmlHttp.status == 200 && callback) {
                callback(JSON.parse(xmlHttp.responseText));
            }
        }
        xmlHttp.open("DELETE", theUrl, true); // true for asynchronous 
        xmlHttp.send(JSON.stringify(requestBody));
    }

    getStartedTime(callback?: ResponseCallback<StartedResponse>): void {
        this.fetchJSON("./started", callback);
    }

    getSchema(callback: ResponseCallback<GraphInstance>): void {
        this.fetchJSON("./schema", callback);
    }

    setParameter(key: string, data, binary: boolean, callback): void {
        const url = "./parameter/value/" + key;
        if (binary) {
            this.postBinaryEmptyResponse(url, data, callback);
        } else {
            this.postJsonBodyEmptyResponse(url, data, callback);
        }
    }

    setParameterTitle(inNodeID: string, value: string, callback?: () => void): void {
        this.postTextBodyEmptyResponse("./parameter/name/" + inNodeID, value, callback);
    }

    setParameterInfo(inNodeID: string, value: string, callback): void {
        this.postTextBodyEmptyResponse("./parameter/description/" + inNodeID, value, callback);
    }

    setProducerTitle(inNodeID: string, value: SetProducerBody, callback): void {
        this.postJsonBodyEmptyResponse("./producer/name/" + inNodeID, value, callback);
    }

    getParameterValue(key: string, callback): void {
        this.fetchRaw("./parameter/value/" + key, callback);
    }

    deleteNodeInput(nodeID: string, inputPortName: string, callback?: ResponseCallback<any>): void {
        this.deleteJSONBodyJSONResponse("node/connection", {
            "nodeId": nodeID,
            "inPortName": inputPortName
        }, callback)
    }

    setNodeInputConnection(
        inNodeID: string,
        inputPortName: string,
        outNodeID: string,
        outPortName: string,
        callback?: ResponseCallback<any>
    ): void {
        this.postJsonBodyJsonResponse("node/connection", {
            "nodeOutId": outNodeID,
            "outPortName": outPortName,
            "nodeInId": inNodeID,
            "inPortName": inputPortName
        }, callback)
    }

    setNodeMetadata(inNodeID: string, key: string, metadata: any, callback?: ResponseCallback<any>): void {
        this.postJsonBodyJsonResponse(`graph/metadata/nodes/${inNodeID}/${key}`, metadata, callback)
    }

    createNote(noteID: string, note, callback?: ResponseCallback<any>): void {
        this.postJsonBodyJsonResponse(`graph/metadata/notes/${noteID}`, note, callback)
    }

    setNoteMetadata(noteID: string, key: string, metadata, callback?: ResponseCallback<any>): void {
        this.postJsonBodyJsonResponse(`graph/metadata/notes/${noteID}/${key}`, metadata, callback)
    }

    deleteMetadata(path: string, callback?): void {
        this.deleteEmptyBodyEmptyResponse(`graph/metadata/${path}`, callback)
    }

    createNode(nodeType: string, callback?: ResponseCallback<CreateNodeResponse>): void {
        this.postJsonBodyJsonResponse("node", {
            "nodeType": nodeType,
        }, callback)
    }

    deleteNode(nodeId: string, callback?: ResponseCallback<any>): void {
        this.deleteJSONBodyJSONResponse("node", {
            "nodeID": nodeId,
        }, callback)
    }

    getNodeTypes(callback?: ResponseCallback<Array<NodeType>>): void {
        this.fetchJSON("/node-types", callback)
    }

    getGraph(callback): void {
        this.fetchJSON("./graph", callback)
    }

    getManifest(nodeId: string, portName: string, callback?: ResponseCallback<Manifest>): void {
        this.fetchJSON(`./manifest/${nodeId}/${portName}`, callback)
    }

    getSwagger(callback): void {
        this.fetchJSON("./swagger", callback)
    }

    setGraph(newGraph, callback?: ResponseCallback<any>): void {
        this.postJsonBodyJsonResponse("./graph", newGraph, callback)
    }
}