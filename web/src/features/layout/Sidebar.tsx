import { GraphControls } from "@/features/graph/GraphControls";
import { ProfileSection } from "@/features/profiles/ProfileSection";
import { VariableSection } from "@/features/variables/VariableSection";
import { SubGraphSection } from "@/features/subgraph/SubGraphSection";
import { useUiStore } from "@/stores/uiStore";
import { ReactNode, useState } from "react";

import { Puzzle, Variable, BookMarked, Video, Workflow } from "lucide-react";
import { RenderingSection } from "../rendering/RenderingSection";
import { ThreeApp } from "@/lib/three_app";
import { NodesSection } from "../NodesSection/NodesSection";

enum SidebarSection {
  Subgraphs = "Subgraphs",
  Profiles = "Profiles",
  Variables = "Variables",
  Rendering = "Rendering",
  Nodes = "Nodes",
}

interface SideBarIconProps {
  icon: ReactNode;
  section: SidebarSection;
  currentState: SidebarSection;
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

  const [sectionShown, setSectionShown] = useState<SidebarSection>(
    SidebarSection.Variables,
  );

  return (
    <div id="sidebar">
      <div id="sidebar-section-icons">
        <SideBarIcon
          section={SidebarSection.Nodes}
          currentState={sectionShown}
          click={setSectionShown}
          icon={<Workflow />}
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
          section={SidebarSection.Profiles}
          currentState={sectionShown}
          click={setSectionShown}
          icon={<BookMarked />}
        />
        <SideBarIcon
          section={SidebarSection.Rendering}
          currentState={sectionShown}
          click={setSectionShown}
          icon={<Video />}
        />
      </div>
      <div id="sidebar-content">
        {!hideInfo && (
          <div id="info">
            <h1>{document.title || "Polyform"}</h1>
          </div>
        )}
        <GraphControls />
        {sectionShown == SidebarSection.Variables && <VariableSection />}
        {sectionShown == SidebarSection.Subgraphs && <SubGraphSection />}
        {sectionShown == SidebarSection.Profiles && <ProfileSection />}
        {sectionShown == SidebarSection.Rendering && <RenderingSection />}
        {sectionShown == SidebarSection.Nodes && <NodesSection />}
      </div>
    </div>
  );
}
