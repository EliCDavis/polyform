import { MessageOverlay } from "@/components/MessageOverlay";
import { PortTypePickerModal } from "@/features/popups/PortTypePickerModal";

/** Global overlays and store-driven modals (imperative APIs from non-React code). */
export function AppOverlays() {
  return (
    <>
      <MessageOverlay />
      <PortTypePickerModal />
    </>
  );
}
