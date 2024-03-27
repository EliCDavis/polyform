class SchemaManager {
    constructor(requestManager, nodeManager) {
        this.requestManager = requestManager;
        this.nodeManager = nodeManager;

        this.schema = null;
        this.subscribers = [];
    }

    subscribe(subscriber) {
        this.subscribers.push(subscriber);
    }

    setProfileKey(key, data, binary) {
        this.requestManager.updateProfile(
            key,
            data,
            binary,
            () => {
                // this.refreshSchema();
            }
        )
    }

    refreshSchema() {
        this.requestManager.getSchema((newSchema) => {
            this.schema = newSchema;
            this.subscribers.forEach(sub => {
                sub(this.schema);
            });

            this.nodeManager.updateNodes(this.schema)
        })
    }
}