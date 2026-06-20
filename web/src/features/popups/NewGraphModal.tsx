import { useState } from "react";
import { Modal, PopupButton } from "@/components/Modal";
import { getExampleGraphs } from "@/config";

interface NewGraphModalProps {
  open: boolean;
  onClose: () => void;
}

export function NewGraphModal({ open, onClose }: NewGraphModalProps) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [author, setAuthor] = useState("");
  const [version, setVersion] = useState("");
  const examples = getExampleGraphs();

  const createGraph = () => {
    fetch("./new-graph", {
      method: "POST",
      body: JSON.stringify({
        name: name || "New Graph",
        author: author || "",
        description: description || "",
        version: version || "v0.0.0",
      }),
    }).then((resp) => {
      if (resp.ok) location.reload();
      else console.error(resp);
    });
  };

  const loadExample = (example: string) => {
    fetch("./load-example", { method: "POST", body: example }).then((resp) => {
      if (resp.ok) location.reload();
      else console.error(resp);
    });
  };

  return (
    <Modal
      title="New Graph"
      open={open}
      onClose={onClose}
      actions={
        <>
          <PopupButton onClick={onClose}>Close</PopupButton>
          <PopupButton variant="primary" onClick={createGraph}>
            New
          </PopupButton>
        </>
      }
    >
      <div style={{ display: "flex", gap: 24 }}>
        <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
          <h3 style={{ fontWeight: "bold", margin: 0 }}>New</h3>
          <label>Name</label>
          <input type="text" value={name} onChange={(e) => setName(e.target.value)} />
          <label>Description</label>
          <textarea value={description} onChange={(e) => setDescription(e.target.value)} />
          <label>Author</label>
          <input type="text" value={author} onChange={(e) => setAuthor(e.target.value)} />
          <label>Version</label>
          <input type="text" value={version} onChange={(e) => setVersion(e.target.value)} />
        </div>
        <div style={{ alignSelf: "center" }}>OR</div>
        <div>
          <h3 style={{ fontWeight: "bold" }}>Open Example</h3>
          <div style={{ width: 170 }}>
            {examples.map((ex) => (
              <button
                key={ex}
                type="button"
                className="example-graph-item"
                onClick={() => loadExample(ex)}
              >
                {ex}
              </button>
            ))}
          </div>
        </div>
      </div>
    </Modal>
  );
}
