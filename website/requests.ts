import { Observable, Subject } from "rxjs";
import { GraphInstance, Manifest, NodeInstance, NodeDefinition, GraphExecutionReport, RegisteredTypes, CreateVariableResponse } from "./schema";

enum GraphChangeEventType {
    // Node
    Node_New = "Node_New",
    Node_Connection = "Node_Connection",
    Node_Metadata = "Node_Metadata",
    Node_Delete = "Node_Delete",

    // Note
    Note_New = "Note_New",
    Note_Metadata = "Note_Metadata",

    // Variable
    Variable_New = "Variable_New",
    Variable_Delete = "Variable_Delete",
    Variable_Info = "Variable_Info",
    Variable_Set = "Variable_Set",

    // Profile
    Profile_New = "Profile_New",
    Profile_Delete = "Profile_Delete",
    Profile_Apply = "Profile_Apply",
    Profile_Rename = "Profile_Rename",
    Profile_Overwrite = "Profile_Overwrite",

    // Graph
    Parameter = "Parameter",
    Producer = "Producer",
    GraphMetadata = "GraphMetadata",
    WholeGraph = "WholeGraph",
}

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

    graphChangeCallbacks: Array<(e: GraphChangeEventType) => void>;

    constructor() {
        this.graphChangeCallbacks = new Array();
    }

    subsribeToGraphChange(cb: (e: GraphChangeEventType) => void): void {
        this.graphChangeCallbacks.push(cb);
    }

    alertGraphHasChanged(e: GraphChangeEventType): void {
        for (let i = 0; i < this.graphChangeCallbacks.length; i++) {
            this.graphChangeCallbacks[i](e);
        }
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

    post$(url: string, body: BodyInit): Observable<Response> {
        const out = new Subject<Response>();
        fetch(url, {
            method: "POST",
            body: body
        }).then((resp) => {
            out.next(resp);
        });
        return out;
    }

    getStartedTime(callback?: ResponseCallback<StartedResponse>): void {
        this.fetchJSON("./started", callback);
    }

    getSchema(callback: ResponseCallback<GraphInstance>): void {
        this.fetchJSON("./schema", callback);
    }

    getExecutionReport(callback: ResponseCallback<GraphExecutionReport>): void {
        this.fetchJSON("./graph/execution-report", callback);
    }

    setParameter(key: string, data, binary: boolean, callback): void {
        const url = "./parameter/value/" + key;
        if (binary) {
            this.postBinaryEmptyResponse(
                url,
                data,
                this.wrapResponseCallbackForGraphChange(GraphChangeEventType.Parameter, callback)
            );
        } else {
            this.postJsonBodyEmptyResponse(
                url,
                data,
                this.wrapResponseCallbackForGraphChange(GraphChangeEventType.Parameter, callback)
            );
        }
    }

    setParameterTitle(inNodeID: string, value: string, callback?: () => void): void {
        this.postTextBodyEmptyResponse(
            "./parameter/name/" + inNodeID,
            value,
            this.wrapCallbackForGraphChange(GraphChangeEventType.Parameter, callback)
        );
    }

    setParameterInfo(inNodeID: string, value: string, callback?: () => void): void {
        this.postTextBodyEmptyResponse(
            "./parameter/description/" + inNodeID,
            value,
            this.wrapCallbackForGraphChange(GraphChangeEventType.Parameter, callback)
        );
    }

    setProducerTitle(inNodeID: string, value: SetProducerBody, callback): void {
        this.postJsonBodyEmptyResponse(
            "./producer/name/" + inNodeID,
            value,
            this.wrapResponseCallbackForGraphChange(GraphChangeEventType.Producer, callback)
        );
    }

    getParameterValue(key: string, callback): void {
        this.fetchRaw("./parameter/value/" + key, callback);
    }

    deleteNodeInput(nodeID: string, inputPortName: string, callback?: ResponseCallback<any>): void {
        this.deleteJSONBodyJSONResponse("node/connection", {
            "nodeId": nodeID,
            "inPortName": inputPortName
        }, this.wrapResponseCallbackForGraphChange(GraphChangeEventType.Node_Connection, callback))
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
        }, this.wrapResponseCallbackForGraphChange(GraphChangeEventType.Node_Connection, callback))
    }

    setNodeMetadata(inNodeID: string, key: string, metadata: any, callback?: ResponseCallback<any>): void {
        this.postJsonBodyJsonResponse(
            `graph/metadata/nodes/${inNodeID}/${key}`,
            metadata,
            this.wrapResponseCallbackForGraphChange(GraphChangeEventType.Node_Metadata, callback)
        );
    }

    deleteNodeMetadata(nodeID: string, callback?): void {
        this.deleteMetadata(
            `nodes/${nodeID}`,
            this.wrapResponseCallbackForGraphChange(GraphChangeEventType.Node_Metadata, callback)
        );
    }

    createNote(noteID: string, note, callback?: ResponseCallback<any>): void {
        this.postJsonBodyJsonResponse(
            `graph/metadata/notes/${noteID}`,
            note,
            this.wrapResponseCallbackForGraphChange(GraphChangeEventType.Note_New, callback)
        );
    }

    setNoteMetadata(noteID: string, key: string, metadata, callback?: ResponseCallback<any>): void {
        this.postJsonBodyJsonResponse(
            `graph/metadata/notes/${noteID}/${key}`,
            metadata,
            this.wrapResponseCallbackForGraphChange(GraphChangeEventType.Note_Metadata, callback)
        )
    }

    deleteMetadata(path: string, callback?): void {
        this.deleteEmptyBodyEmptyResponse(
            `graph/metadata/${path}`,
            this.wrapResponseCallbackForGraphChange(GraphChangeEventType.GraphMetadata, callback)
        )
    }

    createNode(nodeType: string, callback?: ResponseCallback<CreateNodeResponse>): void {
        this.postJsonBodyJsonResponse("node", {
            "nodeType": nodeType,
        }, this.wrapResponseCallbackForGraphChange(GraphChangeEventType.Node_New, callback))
    }

    deleteNode(nodeId: string, callback?: ResponseCallback<any>): void {
        this.deleteNodeMetadata(nodeId);
        this.deleteJSONBodyJSONResponse("node", {
            "nodeID": nodeId,
        }, this.wrapResponseCallbackForGraphChange(GraphChangeEventType.Node_Delete, callback))
    }

    getNodeTypes(callback?: ResponseCallback<RegisteredTypes>): void {
        this.fetchJSON("./node-types", callback)
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
        this.postJsonBodyEmptyResponse(
            "./graph",
            newGraph,
            this.wrapResponseCallbackForGraphChange(GraphChangeEventType.WholeGraph, callback)
        );
    }

    deleteVariable(variableKey: string, success: ((r: Response) => void), error: ((r: Response) => void)): void {
        fetch("./variable/instance/" + variableKey, {
            method: "DELETE",
        }).then((resp) => {
            resp.ok ? success(resp) : error(resp);
            this.alertGraphHasChanged(GraphChangeEventType.Variable_Delete);
        })
    }

    newVariable(variableKey: string, body: any, success: ((r: CreateVariableResponse) => void), error: ((r: any) => void)): void {
        fetch("./variable/instance/" + variableKey, {
            method: "POST",
            body: JSON.stringify(body)
        }).
            then(resp => resp.json()).
            then((resp) => {
                success(resp);
                this.alertGraphHasChanged(GraphChangeEventType.Variable_New);
            }).
            catch(err => {
                error(err)
            });
    }

    updateVariable(variableKey: string, body: any, success: ((r: any) => void), error: ((r: any) => void)): void {
        fetch(
            "./variable/info/" + variableKey, {
            method: "POST",
            body: JSON.stringify(body)
        }).
            then(resp => resp.json()).
            then((resp) => {
                success(resp);
                this.alertGraphHasChanged(GraphChangeEventType.Variable_Info);
            }).
            catch(err => {
                error(err)
            });
    }

    setVariableValue(variable: string, value: any): Observable<Response> {
        return this.post$("./variable/value/" + variable, JSON.stringify(value))

    }

    setBinaryVariableValue(variableKey: string, cb): void {
        const input = document.createElement('input');
        input.type = 'file';

        input.onchange = e => {
            const file = (e.target as HTMLInputElement).files[0];

            const reader = new FileReader();
            reader.readAsArrayBuffer(file);

            reader.onload = readerEvent => {
                const content = readerEvent.target.result as string; // this is the content!
                this.postBinaryEmptyResponse("./variable/value/" + variableKey, content, cb)
            }
        }

        input.click();
    }

    newProfile(profileName: string, success: ((r: any) => void), error: ((r: any) => void)): void {
        fetch("./profile", {
            method: "POST",
            body: JSON.stringify({
                name: profileName,
            })
        }).then((resp) => {
            resp.ok ? success(resp) : error(resp);
            this.alertGraphHasChanged(GraphChangeEventType.Profile_New);
        });
    }

    overwriteProfile(profileName: string, success: ((r: any) => void), error: ((r: any) => void)): void {
        fetch("./profile/overwrite", {
            method: "POST",
            body: JSON.stringify({
                name: profileName,
            })
        }).then((resp) => {
            resp.ok ? success(resp) : error(resp);
            this.alertGraphHasChanged(GraphChangeEventType.Profile_Overwrite);
        });
    }

    deleteProfile(profileName: string, success: ((r: any) => void), error: ((r: any) => void)): void {
        fetch("./profile", {
            method: "DELETE",
            body: JSON.stringify({
                name: profileName,
            })
        }).then((resp) => {
            resp.ok ? success(resp) : error(resp);
            this.alertGraphHasChanged(GraphChangeEventType.Profile_Delete);
        });
    }

    renameProfile(oldName: string, newName: string, success: ((r: any) => void), error: ((r: any) => void)): void {
        fetch("./profile/rename", {
            method: "POST",
            body: JSON.stringify({
                original: oldName,
                new: newName,
            })
        }).then((resp) => {
            resp.ok ? success(resp) : error(resp);
            this.alertGraphHasChanged(GraphChangeEventType.Profile_Rename);
        });
    }

    applyProfile(profileName: string, success: ((r: any) => void), error: ((r: any) => void)): void {
        fetch("./profile/apply", {
            method: "POST",
            body: JSON.stringify({
                name: profileName,
            })
        }).then((resp) => {
            resp.ok ? success(resp) : error(resp);
            this.alertGraphHasChanged(GraphChangeEventType.Profile_Apply);
        });
    }


    private wrapCallbackForGraphChange<T>(e: GraphChangeEventType, callback?: () => void): (() => void) {
        if (callback) {
            return (): void => {
                callback();
                this.alertGraphHasChanged(e);
            };
        }

        return (): void => {
            this.alertGraphHasChanged(e);
        };
    }

    private wrapResponseCallbackForGraphChange<T>(e: GraphChangeEventType, callback?: ResponseCallback<T>): ResponseCallback<T> {
        if (callback) {
            return (responseBody: T): void => {
                callback(responseBody);
                this.alertGraphHasChanged(e);
            };
        }

        return (responseBody: T): void => {
            this.alertGraphHasChanged(e);
        };
    }
}