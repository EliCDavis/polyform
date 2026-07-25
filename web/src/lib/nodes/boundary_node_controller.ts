import { GlobalWidgetFactory, type FlowNode } from "@elicdavis/node-flow";
import type { NodeManager } from "@/lib/node_manager";
import type { RequestManager } from "@/lib/requests";
import { subGraphBoundaryInfo, type NodeInstance } from "@/lib/schema";
import { formatPortTypeLabel } from "@/lib/portTypes";

export class SubGraphBoundaryNodeController {
  private portType: string;
  private suppressTitleSync = false;

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
      if (this.suppressTitleSync || !this.portType) {
        return;
      }
      this.syncBoundaryInfo(newTitle);
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
    const trimmed = portName.trim();
    if (!trimmed || !this.portType) return;

    this.requestManager.setBoundaryNodeInfo(
      this.nodeId,
      { portName: trimmed, scope: this.nodeManager.getScopeApiPath() },
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

    this.portType = boundary.portType;
    this.flowNode.setProperty("portType", boundary.portType);

    if (boundary.portName && boundary.portName !== this.flowNode.title()) {
      this.suppressTitleSync = true;
      this.flowNode.setTitle(boundary.portName);
      this.suppressTitleSync = false;
    }
  }
}
