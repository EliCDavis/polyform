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
      <div className="new-graph">
        <section className="new-graph-form">
          <h3>Blank graph</h3>
          <label className="new-graph-field">
            <span>Name</span>
            <input type="text" placeholder="New Graph" value={name} onChange={(e) => setName(e.target.value)} />
          </label>
          <label className="new-graph-field">
            <span>Description</span>
            <textarea rows={3} value={description} onChange={(e) => setDescription(e.target.value)} />
          </label>
          <label className="new-graph-field">
            <span>Author</span>
            <input type="text" value={author} onChange={(e) => setAuthor(e.target.value)} />
          </label>
          <label className="new-graph-field">
            <span>Version</span>
            <input type="text" placeholder="v0.0.0" value={version} onChange={(e) => setVersion(e.target.value)} />
          </label>
        </section>
        <div className="new-graph-divider">or</div>
        <section className="new-graph-examples">
          <h3>Open an example</h3>
          <div className="new-graph-example-list">
            {examples.map((ex) => (
              <button
                key={ex}
                type="button"
                className="new-graph-example"
                onClick={() => loadExample(ex)}
              >
                {ex.replace(/\.json$/, "")}
              </button>
            ))}
          </div>
        </section>
      </div>
    </Modal>
  );
}
