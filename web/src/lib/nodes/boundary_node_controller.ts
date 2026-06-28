import { GlobalWidgetFactory, type FlowNode } from "@elicdavis/node-flow";
import type { NodeManager } from "@/lib/node_manager";
import type { RequestManager } from "@/lib/requests";
import {
  BoundaryType,
  subGraphBoundaryInfo,
  subGraphBoundaryKind,
  type NodeInstance,
  type RegisteredTypes,
} from "@/lib/schema";
import { formatPortTypeLabel } from "@/lib/portTypes";
import type { BoundaryNodeKind } from "@/features/nodeFlow/subGraphNodeConfigs";
import { portTypePickerActions } from "@/stores/portTypePickerStore";

const CHOOSE_TYPE_LABEL = "Choose type...";

export class SubGraphBoundaryNodeController {
  private portType: string;
  private readonly portTypes: string[];
  private readonly boundaryKind: BoundaryNodeKind;
  private typeButton: ReturnType<typeof GlobalWidgetFactory.create> | null = null;

  constructor(
    private flowNode: FlowNode,
    private readonly nodeManager: NodeManager,
    private readonly requestManager: RequestManager,
    private readonly nodeId: string,
    nodeData: NodeInstance,
    registeredTypes: RegisteredTypes
  ) {
    const boundary = subGraphBoundaryInfo(nodeData);
    this.portTypes = registeredTypes.portTypes ?? [];
    this.portType = boundary?.portType ?? "";
    this.boundaryKind =
      subGraphBoundaryKind(nodeData) === BoundaryType.Input ? "input" : "output";

    if (boundary?.portName) {
      this.flowNode.setTitle(boundary.portName);
    }

    if (this.portType) {
      this.flowNode.setProperty("portType", this.portType);
    }

    this.attachTitleListener();

    this.mountTypeWidgets();

    if (!this.portType) {
      queueMicrotask(() => this.openTypePicker());
    }
  }

  private attachTitleListener(): void {
    this.flowNode.addTitleChangeListener((_, __, newTitle) => {
      if (this.portType) {
        this.syncBoundaryInfo(newTitle, this.portType);
      }
    });
  }

  private mountTypeWidgets(): void {
    const labelWidget = GlobalWidgetFactory.create(this.flowNode, "text", {
      value: "Type",
    });
    this.flowNode.addWidget(labelWidget);
    this.mountTypeButton();
  }

  private typeButtonLabel(): string {
    return this.portType ? formatPortTypeLabel(this.portType) : CHOOSE_TYPE_LABEL;
  }

  private mountTypeButton(): void {
    if (this.typeButton) {
      this.flowNode.removeWidget(this.typeButton);
    }

    const button = GlobalWidgetFactory.create(this.flowNode, "button", {
      text: this.typeButtonLabel(),
      callback: () => this.openTypePicker(),
    });
    this.typeButton = button;
    this.flowNode.addWidget(button);
  }

  private openTypePicker(): void {
    if (this.portTypes.length === 0) return;

    portTypePickerActions.show({
      title: "Port Type",
      options: this.portTypes,
      current: this.portType || this.portTypes[0],
      onSelect: (portType) => {
        this.applyPortType(this.flowNode.title(), portType);
      },
    });
  }

  private applyPortType(portName: string, portType: string): void {
    if (!portType || portType === this.portType) return;

    this.portType = portType;
    this.flowNode = this.nodeManager.replaceBoundaryFlowNode(
      this.nodeId,
      this.flowNode,
      portName,
      portType,
      this.boundaryKind,
    );
    this.flowNode.setProperty("portType", portType);
    this.typeButton = null;
    this.mountTypeWidgets();
    this.attachTitleListener();
    this.syncBoundaryInfo(portName, portType);
  }

  private syncBoundaryInfo(portName: string, portType: string): void {
    if (!portName.trim() || !portType) return;

    this.requestManager.setBoundaryNodeInfo(
      this.nodeId,
      { portName, portType, scope: this.nodeManager.getScopeApiPath() },
      (resp) => {
        this.nodeManager.notifySubGraphDefinitionChanged(resp?.nodeType);
      }
    );
  }

  update(nodeData: NodeInstance): void {
    const boundary = subGraphBoundaryInfo(nodeData);
    if (boundary?.portName) {
      this.flowNode.setTitle(boundary.portName);
    }
    if (boundary?.portType && boundary.portType !== this.portType) {
      this.applyPortType(boundary.portName ?? this.flowNode.title(), boundary.portType);
    }
  }
}
