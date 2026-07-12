import { GlobalWidgetFactory, type FlowNode } from "@elicdavis/node-flow";
import type { NodeManager } from "@/lib/node_manager";
import type { RequestManager } from "@/lib/requests";
import { subGraphBoundaryInfo, type NodeInstance } from "@/lib/schema";
import { formatPortTypeLabel } from "@/lib/portTypes";

export class SubGraphBoundaryNodeController {
  private portType: string;

  constructor(
    private flowNode: FlowNode,
    private readonly nodeManager: NodeManager,
    private readonly requestManager: RequestManager,
    private readonly nodeId: string,
    nodeData: NodeInstance,
  ) {
    const boundary = subGraphBoundaryInfo(nodeData);
    this.portType = boundary?.portType ?? "";

    if (boundary?.portName) {
      this.flowNode.setTitle(boundary.portName);
    }

    if (this.portType) {
      this.flowNode.setProperty("portType", this.portType);
    }

    this.attachTitleListener();
    this.mountTypeWidgets();
  }

  private attachTitleListener(): void {
    this.flowNode.addTitleChangeListener((_, __, newTitle) => {
      if (this.portType) {
        this.syncBoundaryInfo(newTitle);
      }
    });
  }

  private mountTypeWidgets(): void {
    this.flowNode.addWidget(
      GlobalWidgetFactory.create(this.flowNode, "text", {
        value: this.portType ? formatPortTypeLabel(this.portType) : "—",
      }),
    );
  }

  private syncBoundaryInfo(portName: string): void {
    if (!portName.trim() || !this.portType) return;

    this.requestManager.setBoundaryNodeInfo(
      this.nodeId,
      { portName, scope: this.nodeManager.getScopeApiPath() },
      (resp) => {
        this.nodeManager.notifySubGraphDefinitionChanged(resp?.nodeType);
      },
    );
  }

  update(nodeData: NodeInstance): void {
    const boundary = subGraphBoundaryInfo(nodeData);
    if (!boundary) {
      console.error("Boundary node data is missing boundary information", nodeData);
      return;
    }
    this.flowNode.setTitle(boundary.portName);
    this.portType = boundary.portType;
    this.flowNode.setProperty("portType", boundary.portType);
  }
}
