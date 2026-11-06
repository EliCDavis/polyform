import { useEffect, useState } from "react";
import { Move3d } from "lucide-react";
import type { Variable } from "@/types/schema";
import type { ThreeApp } from "@/lib/three_app";
import { DropdownMenu } from "@/components/DropdownMenu";
import { useVariableHoverStore } from "@/stores/variableHoverStore";
import { VariableValueEditor, editorChrome } from "./editors";
import { EditVariableModal } from "@/features/popups/EditVariableModal";
import { DeleteVariableModal } from "@/features/popups/DeleteVariableModal";

interface VariableRowProps {
  variableKey: string;
  label: string;
  variable: Variable;
  threeApp?: ThreeApp;
}

export function VariableRow({ variableKey, label, variable, threeApp }: VariableRowProps) {
  const [editOpen, setEditOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [gizmo, setGizmo] = useState(false);
  const [open, setOpen] = useState(false);
  const hovered = useVariableHoverStore((s) => s.hovered === variableKey);
  const setHovered = useVariableHoverStore((s) => s.setHovered);
  const chrome = editorChrome(variable.type);

  const name = (
    <span className="var-name" title={variableKey}>
      {label}
    </span>
  );
  const menu = (
    <DropdownMenu
      items={[
        { label: "Edit", onClick: () => setEditOpen(true) },
        { label: "Delete", onClick: () => setDeleteOpen(true) },
      ]}
    />
  );
  const description = variable.description && <div className="var-description">{variable.description}</div>;
  const editor = (
    <VariableValueEditor variableKey={variableKey} variable={variable} threeApp={threeApp} gizmo={gizmo} open={open} />
  );

  useEffect(
    () => () => {
      if (useVariableHoverStore.getState().hovered === variableKey) {
        setHovered(null, "sidebar");
      }
    },
    [variableKey, setHovered],
  );

  return (
    <>
      <div
        className={`var-row ${chrome.block ? "var-block" : "var-inline"} ${hovered ? "hovered" : ""}`}
        onPointerEnter={() => setHovered(variableKey, "sidebar")}
        onPointerLeave={() => setHovered(null, "sidebar")}
      >
        {chrome.block ? (
          <>
            <div className="var-header">
              {name}
              {chrome.count && (
                <button type="button" className={`var-count ${open ? "open" : ""}`} aria-expanded={open} onClick={() => setOpen(!open)}>
                  {chrome.count(variable.value)} {chrome.countNoun}
                </button>
              )}
              {chrome.gizmo && threeApp && (
                <button
                  type="button"
                  className={`var-gizmo ${gizmo ? "on" : ""}`}
                  title={gizmo ? "Hide gizmo" : "Show gizmo"}
                  aria-pressed={gizmo}
                  onClick={() => setGizmo(!gizmo)}
                >
                  <Move3d size={14} />
                </button>
              )}
              {menu}
            </div>
            {description}
            {editor}
          </>
        ) : (
          <>
            {name}
            <div className="var-editor">{editor}</div>
            {menu}
            {description}
          </>
        )}
      </div>
      <EditVariableModal
        open={editOpen}
        variableKey={variableKey}
        variable={variable}
        onClose={() => setEditOpen(false)}
      />
      <DeleteVariableModal
        open={deleteOpen}
        variableKey={variableKey}
        onClose={() => setDeleteOpen(false)}
      />
    </>
  );
}
