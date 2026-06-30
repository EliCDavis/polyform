import { Observable, Subject } from "rxjs";
import { RequestManager } from "./requests";
import { GraphInstance } from "./schema";

export class SchemaManager {

    shownPopupOnce: boolean;

    requestManager: RequestManager;

    currentGraph: GraphInstance;

    subscribers: Array<(g: GraphInstance) => void>;

    schema$: Subject<GraphInstance>;

    constructor(requestManager: RequestManager) {
        this.requestManager = requestManager;
        this.schema$ = new Subject<GraphInstance>();

        this.shownPopupOnce = false;
        this.currentGraph = null;
        this.subscribers = [];
    }

    subscribe(subscriber: (g: GraphInstance) => void) {
        this.subscribers.push(subscriber);
    }

    setParameter(key: string, data, binary: boolean) {
        this.requestManager.setParameter(key, data, binary, () => {
            // this.refreshSchema();
        });
    }

    setGraph(newGraph: GraphInstance): void {
        this.currentGraph = newGraph;
        this.subscribers.forEach(sub => {
            sub(this.currentGraph);
        });
        this.schema$.next(this.currentGraph);
    }

    refreshSchema(reason: string): void {
        console.log("Refreshing graph: " + reason)
        this.requestManager.getSchema(this.setGraph.bind(this));
    }

    instance$(): Observable<GraphInstance> {
        return this.schema$.asObservable();
    }
}