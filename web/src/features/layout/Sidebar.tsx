import { FileMenu } from "./FileMenu";
import { VariableSection } from "@/features/variables/VariableSection";
import { SubGraphSection } from "@/features/subgraph/SubGraphSection";
import { useUiStore } from "@/stores/uiStore";
import { ReactNode, useState } from "react";

import { Puzzle, Variable, Video, Search, PanelLeftClose, PanelLeftOpen } from "lucide-react";
import { RenderingSection } from "../rendering/RenderingSection";
import { ThreeApp } from "@/lib/three_app";
import { NodesSection } from "../NodesSection/NodesSection";

enum SidebarSection {
  Subgraphs = "Subgraphs",
  Variables = "Variables",
  Rendering = "Rendering",
  Nodes = "Nodes",
}

interface SideBarIconProps {
  icon: ReactNode;
  section: SidebarSection;
  currentState: SidebarSection | null;
  click: (v: SidebarSection) => void;
}

export function SideBarIcon({
  icon,
  click,
  section,
  currentState,
}: SideBarIconProps) {
  const buttonStyle = {
    backgroundColor: "#003847",
  };

  const selectedStyle = {
    backgroundColor: "#196e6c",
  };

  return (
    <button
      style={currentState == section ? selectedStyle : buttonStyle}
      onClick={() => click(section)}
      title={section}
    >
      {icon}
    </button>
  );
}

interface SidebarProps {}

export function Sidebar({}: SidebarProps) {
  const hideInfo = useUiStore((s) => s.hideInfo);
  const hideFileControls = useUiStore((s) => s.hideFileControls);
  const hideWatermark = useUiStore((s) => s.hideWatermark);
  const graphTitle = useUiStore((s) => s.graphTitle);
  const collapsed = useUiStore((s) => s.sidebarCollapsed);
  const setCollapsed = useUiStore((s) => s.setSidebarCollapsed);
  const drawer = useUiStore((s) => s.compactLayout) && !collapsed;

  const [shown, setShown] = useState<SidebarSection>(
    SidebarSection.Variables,
  );
  const sectionShown = collapsed ? null : shown;

  const setSectionShown = (section: SidebarSection) => {
    if (section === sectionShown) {
      setCollapsed(true);
      return;
    }
    setShown(section);
    setCollapsed(false);
  };

  return (
    <div id="sidebar" className={drawer ? "drawer" : undefined}>
      <div id="sidebar-section-icons">
        <SideBarIcon
          section={SidebarSection.Nodes}
          currentState={sectionShown}
          click={setSectionShown}
          icon={<Search />}
        />
        <SideBarIcon
          section={SidebarSection.Variables}
          currentState={sectionShown}
          click={setSectionShown}
          icon={<Variable />}
        />
        <SideBarIcon
          section={SidebarSection.Subgraphs}
          currentState={sectionShown}
          click={setSectionShown}
          icon={<Puzzle />}
        />
        <SideBarIcon
          section={SidebarSection.Rendering}
          currentState={sectionShown}
          click={setSectionShown}
          icon={<Video />}
        />
        <button
          id="sidebar-collapse"
          onClick={() => setCollapsed(!collapsed)}
          title={collapsed ? "Expand sidebar" : "Collapse sidebar"}
          aria-label={collapsed ? "Expand sidebar" : "Collapse sidebar"}
        >
          {collapsed ? <PanelLeftOpen /> : <PanelLeftClose />}
        </button>
        {!hideWatermark && (
          <a
            id="sidebar-github"
            href="https://github.com/EliCDavis/polyform"
            target="_blank"
            rel="noopener noreferrer"
            title="Polyform on GitHub"
          >
            <i className="fa-brands fa-github" />
          </a>
        )}
      </div>
      <div id="sidebar-content" style={collapsed ? { display: "none" } : undefined}>
        {!hideInfo && (
          <div id="info">
            <h1>{graphTitle}</h1>
            {!hideFileControls && <FileMenu />}
          </div>
        )}
        {sectionShown == SidebarSection.Variables && <VariableSection />}
        {sectionShown == SidebarSection.Subgraphs && <SubGraphSection />}
        {sectionShown == SidebarSection.Rendering && <RenderingSection />}
        {sectionShown == SidebarSection.Nodes && <NodesSection />}
      </div>
    </div>
  );
}
