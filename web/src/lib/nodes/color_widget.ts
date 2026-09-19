import { FlowNode, GlobalWidgetFactory } from '@elicdavis/node-flow';
import type { Widget } from './widget';

export class ColorWidgets {

    private node: FlowNode;

    private listening: Set<string>;

    private updating: boolean;

    constructor(node: FlowNode) {
        this.node = node;
        this.listening = new Set();
        this.updating = false;
    }

    create(property: string, value: string, onChange: (value: string) => void): Widget {
        if (!this.listening.has(property)) {
            this.listening.add(property);
            this.node.addPropertyChangeListener(property, (_old, next) => {
                if (!this.updating) {
                    onChange(String(next));
                }
            });
        }
        const widget = GlobalWidgetFactory.create(this.node, "color", { property });

        // TODO: Fix in Node Flow
        // A fresh widget starts black and only follows the property when it
        // changes, so a rebuild with the same colour has to step through black.
        this.updating = true;
        this.node.setProperty(property, "#000000");
        this.node.setProperty(property, value);
        this.updating = false;
        return widget;
    }
}
