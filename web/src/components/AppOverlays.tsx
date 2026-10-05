import { MessageOverlay } from "@/components/MessageOverlay";
import { PortTypePickerModal } from "@/features/popups/PortTypePickerModal";
import { ConvertToSubGraphModal } from "@/features/popups/ConvertToSubGraphModal";
import { useEditorOptional } from "@/features/editor/EditorContext";

/** Global overlays and store-driven modals (imperative APIs from non-React code). */
export function AppOverlays() {
  const editor = useEditorOptional();
  return (
    <>
      <MessageOverlay />
      <PortTypePickerModal />
      {editor && (
        <ConvertToSubGraphModal
          nodeManager={editor.nodeManager}
          schemaManager={editor.schemaManager}
        />
      )}
    </>
  );
}
