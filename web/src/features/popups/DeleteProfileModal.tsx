import { Modal, PopupButton } from "@/components/Modal";
import type { SchemaManager } from "@/lib/schema_manager";
import { requestManager } from "@/api/client";

interface DeleteProfileModalProps {
  open: boolean;
  profileName: string;
  schemaManager: SchemaManager;
  onClose: () => void;
}

export function DeleteProfileModal({
  open,
  profileName,
  schemaManager,
  onClose,
}: DeleteProfileModalProps) {
  const deleteProfile = () => {
    requestManager.deleteProfile(
      profileName,
      () => {
        schemaManager.refreshSchema("deleted a profile");
        onClose();
      },
      () => alert("unable to delete profile")
    );
  };

  return (
    <Modal
      title="Delete Profile"
      open={open}
      onClose={onClose}
      actions={
        <>
          <PopupButton onClick={onClose}>Close</PopupButton>
          <PopupButton variant="destructive" onClick={deleteProfile}>
            Delete
          </PopupButton>
        </>
      }
    >
      <p>Are you sure you want to delete this profile?</p>
    </Modal>
  );
}
