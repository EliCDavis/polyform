import { useEffect, useState } from "react";
import { Modal, PopupButton } from "@/components/Modal";
import { formatPortTypeLabel } from "@/lib/portTypes";
import { usePortTypePickerStore } from "@/stores/portTypePickerStore";

export function PortTypePickerModal() {
  const request = usePortTypePickerStore((s) => s.request);
  const close = usePortTypePickerStore((s) => s.close);
  const confirm = usePortTypePickerStore((s) => s.confirm);
  const [selected, setSelected] = useState("");

  useEffect(() => {
    if (request) {
      setSelected(request.current);
    }
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
          <PopupButton variant="primary" onClick={() => confirm(selected)}>
            Select
          </PopupButton>
        </>
      }
    >
      <select
        value={selected}
        onChange={(e) => setSelected(e.target.value)}
        style={{ width: "100%" }}
      >
        {request.options.map((type) => (
          <option key={type} value={type}>
            {formatPortTypeLabel(type)}
          </option>
        ))}
      </select>
    </Modal>
  );
}
