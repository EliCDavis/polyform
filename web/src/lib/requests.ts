import type {
  CreateVariableResponse,
  GraphExecutionReport,
  GraphInstance,
  Manifest,
  NodeInstance,
  RegisteredTypes,
} from "./schema";

enum GraphChangeEventType {
  Node_New = "Node_New",
  Node_Connection = "Node_Connection",
  Node_Metadata = "Node_Metadata",
  Node_Delete = "Node_Delete",
  Note_New = "Note_New",
  Note_Metadata = "Note_Metadata",
  Variable_New = "Variable_New",
  Variable_Delete = "Variable_Delete",
  Variable_Info = "Variable_Info",
  Variable_Set = "Variable_Set",
  Profile_New = "Profile_New",
  Profile_Delete = "Profile_Delete",
  Profile_Apply = "Profile_Apply",
  Profile_Rename = "Profile_Rename",
  Profile_Overwrite = "Profile_Overwrite",
  Parameter = "Parameter",
  Producer = "Producer",
  GraphMetadata = "GraphMetadata",
  WholeGraph = "WholeGraph",
}

export type ResponseCallback<T> = (response: T) => void;

export interface SetProducerBody {
  nodePort: string;
  producer: string;
}

export interface StartedResponse {
  time: string;
  modelVersion: number;
}

export interface CreateNodeResponse {
  nodeID: string;
  data: NodeInstance;
}

const JSON_HEADERS = { "Content-Type": "application/json" } as const;

async function getJson<T>(url: string): Promise<T | undefined> {
  try {
    const response = await fetch(url);
    if (!response.ok) return undefined;
    return (await response.json()) as T;
  } catch {
    return undefined;
  }
}

async function getText(url: string): Promise<string | undefined> {
  try {
    const response = await fetch(url);
    if (!response.ok) return undefined;
    return response.text();
  } catch {
    return undefined;
  }
}

async function getBlob(url: string): Promise<Blob | undefined> {
  try {
    const response = await fetch(url);
    if (!response.ok) return undefined;
    return response.blob();
  } catch {
    return undefined;
  }
}

async function postJson<T>(url: string, body: unknown): Promise<T | undefined> {
  try {
    const response = await fetch(url, {
      method: "POST",
      headers: JSON_HEADERS,
      body: JSON.stringify(body),
    });
    if (!response.ok) return undefined;
    const text = await response.text();
    if (!text) return undefined;
    return JSON.parse(text) as T;
  } catch {
    return undefined;
  }
}

async function postJsonVoid(url: string, body: unknown): Promise<boolean> {
  try {
    const response = await fetch(url, {
      method: "POST",
      headers: JSON_HEADERS,
      body: JSON.stringify(body),
    });
    return response.ok;
  } catch {
    return false;
  }
}

async function postTextVoid(url: string, body: string): Promise<boolean> {
  try {
    const response = await fetch(url, { method: "POST", body });
    return response.ok;
  } catch {
    return false;
  }
}

async function postBinaryVoid(url: string, body: BodyInit): Promise<boolean> {
  try {
    const response = await fetch(url, { method: "POST", body });
    return response.ok;
  } catch {
    return false;
  }
}

async function deleteJson<T>(url: string, body: unknown): Promise<T | undefined> {
  try {
    const response = await fetch(url, {
      method: "DELETE",
      headers: JSON_HEADERS,
      body: JSON.stringify(body),
    });
    if (!response.ok) return undefined;
    const text = await response.text();
    if (!text) return undefined;
    return JSON.parse(text) as T;
  } catch {
    return undefined;
  }
}

async function deleteVoid(url: string): Promise<boolean> {
  try {
    const response = await fetch(url, { method: "DELETE" });
    return response.ok;
  } catch {
    return false;
  }
}

export function downloadBlob(url: string, callback: (body: Blob) => void): void {
  void getBlob(url).then((blob) => {
    if (blob) callback(blob);
  });
}

export function saveFileToDisk(url: string, fileName: string): void {
  downloadBlob(url, (data) => {
    const anchor = document.createElement("a");
    anchor.download = fileName;
    const objectUrl = window.URL.createObjectURL(data);
    anchor.href = objectUrl;
    anchor.click();
    window.URL.revokeObjectURL(objectUrl);
  });
}

export class RequestManager {
  private graphChangeListeners: Array<(event: GraphChangeEventType) => void> = [];

  subscribeToGraphChange(listener: (event: GraphChangeEventType) => void): void {
    this.graphChangeListeners.push(listener);
  }

  private notifyGraphChange(event: GraphChangeEventType): void {
    for (const listener of this.graphChangeListeners) {
      listener(event);
    }
  }

  private onGraphChange(event: GraphChangeEventType, callback?: () => void): () => void {
    return () => {
      callback?.();
      this.notifyGraphChange(event);
    };
  }

  private onGraphChangeResponse<T>(
    event: GraphChangeEventType,
    callback?: ResponseCallback<T>
  ): ResponseCallback<T> {
    return (response: T) => {
      callback?.(response);
      this.notifyGraphChange(event);
    };
  }

  fetchImage(
    imgUrl: string,
    successCallback: (img: HTMLImageElement) => void,
    errorCallback: (event: Event | string) => void
  ): void {
    const img = document.createElement("img");
    img.src = `${imgUrl}?${performance.now()}`;
    img.onload = () => successCallback(img);
    img.onerror = (event) => errorCallback(event);
  }

  fetchText(
    url: string,
    successCallback: (text: string) => void,
    errorCallback?: (text: string) => void
  ): void {
    void getText(url).then((text) => {
      if (text !== undefined) successCallback(text);
      else if (errorCallback) errorCallback("");
    });
  }

  fetchJSON<T>(url: string, callback: ResponseCallback<T>): void {
    void getJson<T>(url).then((data) => {
      if (data !== undefined) callback(data);
    });
  }

  fetchRaw(url: string, callback: (blob: Blob) => void): void {
    void getBlob(url).then((blob) => {
      if (blob) callback(blob);
    });
  }

  postBinaryEmptyResponse(url: string, body: BodyInit, callback?: () => void): void {
    void postBinaryVoid(url, body).then((ok) => {
      if (ok) callback?.();
    });
  }

  getStartedTime(callback?: ResponseCallback<StartedResponse>): void {
    this.fetchJSON("./started", callback ?? (() => {}));
  }

  getSchema(callback: ResponseCallback<GraphInstance>): void {
    this.fetchJSON("./schema", callback);
  }

  getExecutionReport(callback: ResponseCallback<GraphExecutionReport>): void {
    this.fetchJSON("./graph/execution-report", callback);
  }

  setParameter(key: string, data: unknown, binary: boolean, callback?: () => void): void {
    const url = `./parameter/value/${key}`;
    const wrapped = this.onGraphChange(GraphChangeEventType.Parameter, callback);
    if (binary) {
      void postBinaryVoid(url, data as BodyInit).then((ok) => {
        if (ok) wrapped();
      });
    } else {
      void postJsonVoid(url, data).then((ok) => {
        if (ok) wrapped();
      });
    }
  }

  setParameterTitle(nodeId: string, value: string, callback?: () => void): void {
    void postTextVoid(
      `./parameter/name/${nodeId}`,
      value,
    ).then((ok) => {
      if (ok) this.onGraphChange(GraphChangeEventType.Parameter, callback)();
    });
  }

  setParameterInfo(nodeId: string, value: string, callback?: () => void): void {
    void postTextVoid(
      `./parameter/description/${nodeId}`,
      value,
    ).then((ok) => {
      if (ok) this.onGraphChange(GraphChangeEventType.Parameter, callback)();
    });
  }

  setProducerTitle(nodeId: string, value: SetProducerBody, callback?: () => void): void {
    void postJsonVoid(`./producer/name/${nodeId}`, value).then((ok) => {
      if (ok) this.onGraphChange(GraphChangeEventType.Producer, callback)();
    });
  }

  getParameterValue(key: string, callback: (blob: Blob) => void): void {
    this.fetchRaw(`./parameter/value/${key}`, callback);
  }

  deleteNodeInput(nodeId: string, inputPortName: string, callback?: ResponseCallback<unknown>): void {
    void deleteJson("node/connection", {
      nodeId,
      inPortName: inputPortName,
    }).then((response) => {
      if (response !== undefined) {
        this.onGraphChangeResponse(GraphChangeEventType.Node_Connection, callback)(response);
      }
    });
  }

  setNodeInputConnection(
    inNodeId: string,
    inputPortName: string,
    outNodeId: string,
    outPortName: string,
    callback?: ResponseCallback<unknown>
  ): void {
    void postJson("node/connection", {
      nodeOutId: outNodeId,
      outPortName,
      nodeInId: inNodeId,
      inPortName: inputPortName,
    }).then((response) => {
      if (response !== undefined) {
        this.onGraphChangeResponse(GraphChangeEventType.Node_Connection, callback)(response);
      }
    });
  }

  setNodeMetadata(
    nodeId: string,
    key: string,
    metadata: unknown,
    callback?: ResponseCallback<unknown>
  ): void {
    void postJson(`graph/metadata/nodes/${nodeId}/${key}`, metadata).then((response) => {
      if (response !== undefined) {
        this.onGraphChangeResponse(GraphChangeEventType.Node_Metadata, callback)(response);
      }
    });
  }

  deleteNodeMetadata(nodeId: string, callback?: () => void): void {
    this.deleteMetadata(`nodes/${nodeId}`, callback);
  }

  createNote(noteId: string, note: unknown, callback?: ResponseCallback<unknown>): void {
    void postJson(`graph/metadata/notes/${noteId}`, note).then((response) => {
      if (response !== undefined) {
        this.onGraphChangeResponse(GraphChangeEventType.Note_New, callback)(response);
      }
    });
  }

  setNoteMetadata(
    noteId: string,
    key: string,
    metadata: unknown,
    callback?: ResponseCallback<unknown>
  ): void {
    void postJson(`graph/metadata/notes/${noteId}/${key}`, metadata).then((response) => {
      if (response !== undefined) {
        this.onGraphChangeResponse(GraphChangeEventType.Note_Metadata, callback)(response);
      }
    });
  }

  deleteMetadata(path: string, callback?: () => void): void {
    void deleteVoid(`graph/metadata/${path}`).then((ok) => {
      if (ok) this.onGraphChange(GraphChangeEventType.GraphMetadata, callback)();
    });
  }

  createNode(nodeType: string, callback?: ResponseCallback<CreateNodeResponse>): void {
    void postJson<CreateNodeResponse>("node", { nodeType }).then((response) => {
      if (response !== undefined) {
        this.onGraphChangeResponse(GraphChangeEventType.Node_New, callback)(response);
      }
    });
  }

  deleteNode(nodeId: string, callback?: ResponseCallback<unknown>): void {
    this.deleteNodeMetadata(nodeId);
    void deleteJson("node", { nodeID: nodeId }).then((response) => {
      if (response !== undefined) {
        this.onGraphChangeResponse(GraphChangeEventType.Node_Delete, callback)(response);
      }
    });
  }

  getNodeTypes(callback?: ResponseCallback<RegisteredTypes>): void {
    this.fetchJSON("./node-types", callback ?? (() => {}));
  }

  getGraph(callback: ResponseCallback<GraphInstance>): void {
    this.fetchJSON("./graph", callback);
  }

  getManifest(nodeId: string, portName: string, callback?: ResponseCallback<Manifest>): void {
    this.fetchJSON(`./manifest/${nodeId}/${portName}`, callback ?? (() => {}));
  }

  getSwagger(callback: ResponseCallback<unknown>): void {
    this.fetchJSON("./swagger", callback);
  }

  setGraph(newGraph: unknown, callback?: () => void): void {
    void postJsonVoid("./graph", newGraph).then((ok) => {
      if (ok) this.onGraphChange(GraphChangeEventType.WholeGraph, callback)();
    });
  }

  deleteVariable(
    variableKey: string,
    success: (response: Response) => void,
    error: (response: Response) => void
  ): void {
    void fetch(`./variable/instance/${variableKey}`, { method: "DELETE" }).then((response) => {
      response.ok ? success(response) : error(response);
      this.notifyGraphChange(GraphChangeEventType.Variable_Delete);
    });
  }

  newVariable(
    variableKey: string,
    body: { type: string; description?: string },
    success: (response: CreateVariableResponse) => void,
    error: (err: unknown) => void
  ): void {
    void fetch(`./variable/instance/${variableKey}`, {
      method: "POST",
      headers: JSON_HEADERS,
      body: JSON.stringify(body),
    })
      .then((response) => response.json())
      .then((response) => {
        success(response as CreateVariableResponse);
        this.notifyGraphChange(GraphChangeEventType.Variable_New);
      })
      .catch(error);
  }

  updateVariable(
    variableKey: string,
    body: { name: string; description?: string },
    success: (response: unknown) => void,
    error: (err: unknown) => void
  ): void {
    void fetch(`./variable/info/${variableKey}`, {
      method: "POST",
      headers: JSON_HEADERS,
      body: JSON.stringify(body),
    })
      .then((response) => response.json())
      .then((response) => {
        success(response);
        this.notifyGraphChange(GraphChangeEventType.Variable_Info);
      })
      .catch(error);
  }

  newProfile(
    profileName: string,
    success: (response: Response) => void,
    error: (response: Response) => void
  ): void {
    this.profileRequest("./profile", GraphChangeEventType.Profile_New, { name: profileName }, success, error);
  }

  overwriteProfile(
    profileName: string,
    success: (response: Response) => void,
    error: (response: Response) => void
  ): void {
    this.profileRequest(
      "./profile/overwrite",
      GraphChangeEventType.Profile_Overwrite,
      { name: profileName },
      success,
      error
    );
  }

  deleteProfile(
    profileName: string,
    success: (response: Response) => void,
    error: (response: Response) => void
  ): void {
    void fetch("./profile", {
      method: "DELETE",
      headers: JSON_HEADERS,
      body: JSON.stringify({ name: profileName }),
    }).then((response) => {
      response.ok ? success(response) : error(response);
      this.notifyGraphChange(GraphChangeEventType.Profile_Delete);
    });
  }

  renameProfile(
    oldName: string,
    newName: string,
    success: (response: Response) => void,
    error: (response: Response) => void
  ): void {
    this.profileRequest(
      "./profile/rename",
      GraphChangeEventType.Profile_Rename,
      { original: oldName, new: newName },
      success,
      error
    );
  }

  applyProfile(
    profileName: string,
    success: (response: Response) => void,
    error: (response: Response) => void
  ): void {
    this.profileRequest(
      "./profile/apply",
      GraphChangeEventType.Profile_Apply,
      { name: profileName },
      success,
      error
    );
  }

  private profileRequest(
    url: string,
    event: GraphChangeEventType,
    body: unknown,
    success: (response: Response) => void,
    error: (response: Response) => void
  ): void {
    void fetch(url, {
      method: "POST",
      headers: JSON_HEADERS,
      body: JSON.stringify(body),
    }).then((response) => {
      response.ok ? success(response) : error(response);
      this.notifyGraphChange(event);
    });
  }
}
