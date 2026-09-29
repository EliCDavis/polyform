import { FlowNode, GlobalWidgetFactory, type Widget } from '@elicdavis/node-flow';

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
        this.updating = true;
        this.node.setProperty(property, value);
        this.updating = false;

        return GlobalWidgetFactory.create(this.node, "color", { property });
    }
}
