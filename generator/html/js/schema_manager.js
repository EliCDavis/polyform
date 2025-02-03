class SchemaManager {
    constructor(requestManager, nodeManager, noteManager) {
        this.modelVersion = -1;
        this.requestManager = requestManager;
        this.nodeManager = nodeManager;
        this.noteManager = noteManager;

        this.schema = null;
        this.subscribers = [];
    }

    subscribe(subscriber) {
        this.subscribers.push(subscriber);
    }

    setParameter(key, data, binary) {
        this.requestManager.setParameter(
            key,
            data,
            binary,
            () => {
                // this.refreshSchema();
            }
        )
    }

    setModelVersion(newModelVersion) {
        if (newModelVersion === this.modelVersion) {
            return;
        }
        this.modelVersion = newModelVersion;
        this.refreshSchema();
    }

    refreshSchema() {
        this.requestManager.getSchema((newSchema) => {
            this.schema = newSchema;
            this.subscribers.forEach(sub => {
                sub(this.schema);
            });

            this.nodeManager.updateNodes(this.schema)
            this.noteManager.schemaUpdate(this.schema);
        })
    }
}