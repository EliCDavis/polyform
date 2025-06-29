import { Observable, Subject } from "rxjs";
import { NodeManager } from "./node_manager";
import { NoteManager } from "./note_manager";
import { NewGraphPopup } from "./popups/new_graph";
import { RequestManager } from "./requests";
import { GraphInstance } from "./schema";

export class SchemaManager {

    modelVersion: number;

    shownPopupOnce: boolean;

    requestManager: RequestManager;

    nodeManager: NodeManager;

    noteManager: NoteManager;

    currentGraph: GraphInstance;

    subscribers: Array<(g: GraphInstance) => void>;

    schema$: Subject<GraphInstance>;

    constructor(requestManager: RequestManager, nodeManager: NodeManager, noteManager: NoteManager) {
        this.modelVersion = -1;
        this.requestManager = requestManager;
        this.nodeManager = nodeManager;
        this.noteManager = noteManager;
        this.schema$ = new Subject<GraphInstance>();

        this.shownPopupOnce = false;
        this.currentGraph = null;
        this.subscribers = [];
    }

    subscribe(subscriber: (g: GraphInstance) => void) {
        this.subscribers.push(subscriber);
    }

    setParameter(key: string, data, binary) {
        this.requestManager.setParameter(
            key,
            data,
            binary,
            () => {
                // this.refreshSchema();
            }
        )
    }

    setModelVersion(newModelVersion: number): void {
        if (newModelVersion === this.modelVersion) {
            return;
        }
        this.modelVersion = newModelVersion;
        this.refreshSchema();
    }

    setGraph(newGraph: GraphInstance): void {
        this.currentGraph = newGraph;
        this.subscribers.forEach(sub => {
            sub(this.currentGraph);
        });
        this.schema$.next(this.currentGraph);

        this.nodeManager.updateNodes(this.currentGraph)
        this.noteManager.schemaUpdate(this.currentGraph);
    }

    refreshSchema(): void {
        this.requestManager.getSchema(this.setGraph.bind(this));
    }

    instance$(): Observable<GraphInstance> {
        return this.schema$.asObservable();
    }
}