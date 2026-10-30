import { useEffect, useRef } from "react";
import { useSchema } from "@/api/hooks";
import { requestManager } from "@/api/client";
import { useEditorOptional } from "@/features/editor/EditorContext";
import { useProfileStore } from "@/stores/profileStore";
import { useUiStore } from "@/stores/uiStore";

export const DEFAULT_PROFILE = "default";

// Keeps one profile open at all times; variable writes name it in their URL.
// On load the remembered profile, or else the first, is applied over whatever
// values the graph was saved with. Later the schema list can lag a profile
// the user just made or renamed, so an active name missing from it is left
// alone; only an empty choice falls back.
export function ActiveProfileSync() {
  const editor = useEditorOptional();
  const { data: graph } = useSchema();
  const canEdit = useUiStore((s) => s.canEdit);
  const active = useProfileStore((s) => s.activeProfile);
  const setActive = useProfileStore((s) => s.setActiveProfile);
  const opened = useRef(false);

  useEffect(() => {
    if (!graph || !editor) return;
    const profiles = graph.profiles ?? [];
    if (profiles.length > 0) {
      if (opened.current) {
        if (active === null) setActive(profiles[0]);
        return;
      }
      opened.current = true;
      const chosen = active !== null && profiles.includes(active) ? active : profiles[0];
      if (!canEdit) {
        setActive(chosen);
        return;
      }
      requestManager.applyProfile(
        chosen,
        () => {
          setActive(chosen);
          editor.schemaManager.refreshSchema("opened a profile");
        },
        () => alert(`unable to open profile "${chosen}"`)
      );
      return;
    }
    if (!canEdit || opened.current) return;
    opened.current = true;
    requestManager.newProfile(
      DEFAULT_PROFILE,
      () => {
        setActive(DEFAULT_PROFILE);
        editor.schemaManager.refreshSchema("created the default profile");
      },
      () => {
        opened.current = false;
      }
    );
  }, [graph, editor, canEdit, active, setActive]);

  return null;
}
