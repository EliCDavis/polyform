class SchemaManager {
    constructor(requestManager, nodeManager, noteManager) {
        this.modelVersion = -1;
        this.requestManager = requestManager;
        this.nodeManager = nodeManager;
        this.noteManager = noteManager;

        this.shownPopupOnce = false;
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
        this.requestManager.getSchema(((newSchema) => {
            console.log(newSchema)

            if (Object.keys(newSchema.nodes).length === 0 && !this.shownPopupOnce) {
                document.getElementById("new-graph-popup").style.display = "flex"; 
                this.shownPopupOnce = true;
            }

            this.schema = newSchema;
            this.subscribers.forEach(sub => {
                sub(this.schema);
            });

            this.nodeManager.updateNodes(this.schema)
            this.noteManager.schemaUpdate(this.schema);
        }).bind(this))
    }
}