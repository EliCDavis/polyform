import { FlowNode, GlobalWidgetFactory } from '@elicdavis/node-flow';
import type { Widget } from './widget';
import { NodeManager } from '../node_manager';
import { ColorWidgets } from './color_widget';
import type { ColorGradientNodeParameter, GradientKey } from '../../types/parameter';

export class GradientParameterNodeController {

    flowNode: FlowNode;

    nodeManager: NodeManager;

    id: string;

    keys: Array<GradientKey>;

    widgets: Array<Widget>;

    colors: ColorWidgets;

    constructor(flowNode: FlowNode, nodeManager: NodeManager, id: string, parameterData: ColorGradientNodeParameter) {
        this.id = id;
        this.nodeManager = nodeManager;
        this.keys = parameterData.currentValue?.keys ?? [];
        this.widgets = [];
        this.colors = new ColorWidgets(flowNode);

        this.flowNode = flowNode;
        this.flowNode.setTitle(parameterData.name);

        this.rebuildWidgets();
    }

    private rebuildWidgets(): void {
        for (const w of this.widgets) {
            this.flowNode.removeWidget(w);
        }
        this.widgets = [];

        const pinned = (i: number) => this.keys.length > 1 && (i === 0 || i === this.keys.length - 1);

        this.keys.forEach((key, index) => {
            this.add(this.colors.create(`stop-${index}`, key.value, (value) =>
                this.commit(this.keys.map((k, i) => (i === index ? { ...k, value } : k)))));
            if (pinned(index)) {
                this.add(GlobalWidgetFactory.create(this.flowNode, "text", { value: `t = ${key.time}` }));
                return;
            }
            this.add(GlobalWidgetFactory.create(this.flowNode, "number", {
                value: key.time,
                callback: (time: number) => this.commit(this.keys.map((k, i) => (i === index ? { ...k, time } : k))),
            }));
        });

        this.add(GlobalWidgetFactory.create(this.flowNode, "button", {
            text: "Add Stop",
            callback: () => {
                const last = this.keys[this.keys.length - 1];
                const previous = this.keys[this.keys.length - 2] ?? last;
                const inserted = last === undefined
                    ? [{ time: 0, value: "#000000" }, { time: 1, value: "#ffffff" }]
                    : [{ time: (previous.time + last.time) / 2, value: last.value }];
                this.commit([...this.keys.slice(0, -1), ...inserted, ...(last ? [last] : [])]);
            },
        }));
        this.add(GlobalWidgetFactory.create(this.flowNode, "button", {
            text: "Remove Stop",
            callback: () => this.commit(this.keys.slice(0, -1)),
        }));
    }

    private add(widget: Widget): void {
        this.widgets.push(widget);
        this.flowNode.addWidget(widget);
    }

    private commit(keys: Array<GradientKey>): void {
        this.keys = [...keys].sort((a, b) => a.time - b.time);
        this.rebuildWidgets();
        this.nodeManager.nodeParameterChanged({
            id: this.id,
            data: { keys: this.keys },
            binary: false
        });
    }

    update(parameterData: ColorGradientNodeParameter): void {
        this.keys = parameterData.currentValue?.keys ?? [];
        this.rebuildWidgets();
    }
}
