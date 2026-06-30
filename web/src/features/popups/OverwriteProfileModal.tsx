import { Modal, PopupButton } from "@/components/Modal";
import type { SchemaManager } from "@/lib/schema_manager";
import { requestManager } from "@/api/client";

interface OverwriteProfileModalProps {
  open: boolean;
  profileName: string;
  schemaManager: SchemaManager;
  onClose: () => void;
}

export function OverwriteProfileModal({
  open,
  profileName,
  schemaManager,
  onClose,
}: OverwriteProfileModalProps) {
  const overwrite = () => {
    requestManager.overwriteProfile(
      profileName,
      () => {
        schemaManager.refreshSchema("profile overwritten");
        onClose();
      },
      () => alert("unable to overwrite profile")
    );
  };

  return (
    <Modal
      title="Overwrite Profile"
      open={open}
      onClose={onClose}
      actions={
        <>
          <PopupButton onClick={onClose}>Close</PopupButton>
          <PopupButton variant="primary" onClick={overwrite}>
            Overwrite
          </PopupButton>
        </>
      }
    >
      <p>
        Are you sure you want to overwrite &quot;{profileName}&quot; with the current state of the
        graph?
      </p>
    </Modal>
  );
}
