class SchemaManager {
    constructor(requestManager, nodeManager) {
        this.requestManager = requestManager;
        this.nodeManager = nodeManager;

        this.schema = null;
        this.profile = {}
        this.subscribers = [];
    }

    subscribe(subscriber) {
        this.subscribers.push(subscriber);
    }

    setProfileKey(key, data) {
        this.profile[key] = data;
    }

    submitProfile() {
        this.requestManager.updateProfile(this.profile, () => {
            // this.refreshSchema();
        })
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