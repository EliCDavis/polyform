import { GraphControls } from "@/features/graph/GraphControls";
import { ProfileSection } from "@/features/profiles/ProfileSection";
import { VariableSection } from "@/features/variables/VariableSection";
import { SubGraphSection } from "@/features/subgraph/SubGraphSection";
import { useUiStore } from "@/stores/uiStore";

export function Sidebar() {
  const hideInfo = useUiStore((s) => s.hideInfo);

  return (
    <div id="sidebar">
      <div id="sidebar-content">
        {!hideInfo && (
          <div id="info">
            <h1>{document.title || "Polyform"}</h1>
          </div>
        )}
        <GraphControls />
        <SubGraphSection />
        <ProfileSection />
        <VariableSection />
      </div>
    </div>
  );
}
