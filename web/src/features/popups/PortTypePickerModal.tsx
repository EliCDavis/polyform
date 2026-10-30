import { useEffect, useState } from "react";
import { Modal, PopupButton } from "@/components/Modal";
import { usePortTypePickerStore } from "@/stores/portTypePickerStore";
import { PortTypeList } from "./PortTypeList";

export function PortTypePickerModal() {
  const request = usePortTypePickerStore((s) => s.request);
  const close = usePortTypePickerStore((s) => s.close);
  const confirm = usePortTypePickerStore((s) => s.confirm);
  const [selected, setSelected] = useState("");

  useEffect(() => {
    if (request) setSelected(request.current);
  }, [request]);

  if (!request) return null;

  return (
    <Modal
      title={request.title}
      open
      onClose={close}
      actions={
        <>
          <PopupButton onClick={close}>Cancel</PopupButton>
          <PopupButton variant="primary" onClick={() => confirm(selected)} disabled={!selected}>
            Select
          </PopupButton>
        </>
      }
    >
      <PortTypeList
        key={request.title + request.current}
        options={request.options}
        selected={selected}
        onSelect={setSelected}
        onConfirm={confirm}
        onCancel={close}
        autoFocus
      />
    </Modal>
  );
}
