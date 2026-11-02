import { FlowNode, GlobalWidgetFactory, ImageWidget, type Widget } from '@elicdavis/node-flow';
import type { GraphVersion } from '../schema.js';

interface OutputPreview {
    image?: ImageWidget;
    button?: Widget;
    stamp?: string;
}

export class OutputPreviews {
    private readonly previews = new Map<string, OutputPreview>();

    constructor(
        private readonly flowNode: FlowNode,
        private readonly graphVersion: () => GraphVersion,
    ) { }

    claim(outputKey: string, version: number): string | null {
        const stamp = `${this.graphVersion().rebuilds}:${Math.max(version, 0)}`;
        const preview = this.entry(outputKey);
        if (preview.stamp === stamp) {
            return null;
        }
        preview.stamp = stamp;
        return stamp;
    }

    isCurrent(outputKey: string, stamp: string): boolean {
        return this.previews.get(outputKey)?.stamp === stamp;
    }

    image(outputKey: string): ImageWidget | undefined {
        return this.previews.get(outputKey)?.image;
    }

    addImage(outputKey: string, config: object = {}): ImageWidget {
        const preview = this.entry(outputKey);
        if (!preview.image) {
            preview.image = GlobalWidgetFactory.create(this.flowNode, "image", config) as ImageWidget;
            this.flowNode.addWidget(preview.image);
        }
        return preview.image;
    }

    hasButton(outputKey: string): boolean {
        return this.previews.get(outputKey)?.button !== undefined;
    }

    addButton(outputKey: string, config: object): void {
        const preview = this.entry(outputKey);
        if (!preview.button) {
            preview.button = GlobalWidgetFactory.create(this.flowNode, "button", config);
            this.flowNode.addWidget(preview.button);
        }
    }

    clear(outputKey: string): void {
        const preview = this.previews.get(outputKey);
        if (!preview) {
            return;
        }
        if (preview.image) {
            this.flowNode.removeWidget(preview.image);
            preview.image = undefined;
        }
        if (preview.button) {
            this.flowNode.removeWidget(preview.button);
            preview.button = undefined;
        }
    }

    remove(outputKey: string): void {
        this.clear(outputKey);
        this.previews.delete(outputKey);
    }

    private entry(outputKey: string): OutputPreview {
        let preview = this.previews.get(outputKey);
        if (!preview) {
            preview = {};
            this.previews.set(outputKey, preview);
        }
        return preview;
    }
}
