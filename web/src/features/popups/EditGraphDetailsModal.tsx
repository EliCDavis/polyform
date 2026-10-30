import { useEffect, useState } from "react";
import { Modal, PopupButton } from "@/components/Modal";
import { requestManager } from "@/api/client";
import type { Author, AuthorContact, GraphDetails } from "@/lib/requests";
import { useUiStore } from "@/stores/uiStore";

interface EditGraphDetailsModalProps {
  open: boolean;
  onClose: () => void;
}

const row = { display: "flex", gap: 8, alignItems: "center" } as const;

export function EditGraphDetailsModal({ open, onClose }: EditGraphDetailsModalProps) {
  const setGraphTitle = useUiStore((s) => s.setGraphTitle);
  const [details, setDetails] = useState<GraphDetails | null>(null);

  useEffect(() => {
    if (!open) return;
    setDetails(null);
    requestManager.getGraphDetails(setDetails);
  }, [open]);

  const authors = details?.authors ?? [];
  const update = (patch: Partial<GraphDetails>) => setDetails((d) => (d ? { ...d, ...patch } : d));
  const setAuthors = (next: Array<Author>) => update({ authors: next });
  const setAuthor = (i: number, author: Author) => setAuthors(authors.map((a, j) => (j === i ? author : a)));
  const setContacts = (i: number, contacts: Array<AuthorContact>) => setAuthor(i, { ...authors[i], contactInfo: contacts });

  const save = () => {
    if (!details) return;
    requestManager.setGraphDetails(details, (saved) => {
      setGraphTitle(saved.name);
      onClose();
    });
  };

  return (
    <Modal
      title="Graph Details"
      open={open}
      onClose={onClose}
      actions={
        <>
          <PopupButton onClick={onClose}>Cancel</PopupButton>
          <PopupButton variant="primary" onClick={save} disabled={!details}>
            Save
          </PopupButton>
        </>
      }
    >
      {!details ? (
        <span>Loading…</span>
      ) : (
        <div style={{ display: "flex", flexDirection: "column", gap: 12, width: 440 }}>
          <label>Name</label>
          <input type="text" value={details.name} onChange={(e) => update({ name: e.target.value })} />
          <label>Version</label>
          <input type="text" value={details.version} onChange={(e) => update({ version: e.target.value })} />
          <label>Description</label>
          <textarea value={details.description} onChange={(e) => update({ description: e.target.value })} />

          <label>Authors</label>
          {authors.map((author, i) => {
            const contacts = author.contactInfo ?? [];
            return (
              <div key={i} style={{ display: "flex", flexDirection: "column", gap: 6, paddingLeft: 8, borderLeft: "2px solid rgba(255,255,255,0.2)" }}>
                <div style={row}>
                  <input
                    type="text"
                    placeholder="Name"
                    value={author.name}
                    onChange={(e) => setAuthor(i, { ...author, name: e.target.value })}
                  />
                  <button type="button" className="secondary" onClick={() => setAuthors(authors.filter((_, j) => j !== i))}>
                    Remove
                  </button>
                </div>
                {contacts.map((contact, c) => (
                  <div key={c} style={row}>
                    <input
                      type="text"
                      placeholder="Where (github.com)"
                      value={contact.medium}
                      onChange={(e) => setContacts(i, contacts.map((x, k) => (k === c ? { ...x, medium: e.target.value } : x)))}
                    />
                    <input
                      type="text"
                      placeholder="Handle"
                      value={contact.value}
                      onChange={(e) => setContacts(i, contacts.map((x, k) => (k === c ? { ...x, value: e.target.value } : x)))}
                    />
                    <button type="button" className="secondary" onClick={() => setContacts(i, contacts.filter((_, k) => k !== c))}>
                      Remove
                    </button>
                  </div>
                ))}
                <button
                  type="button"
                  className="secondary"
                  style={{ alignSelf: "flex-start" }}
                  onClick={() => setContacts(i, [...contacts, { medium: "", value: "" }])}
                >
                  Add contact
                </button>
              </div>
            );
          })}
          <button
            type="button"
            className="secondary"
            style={{ alignSelf: "flex-start" }}
            onClick={() => setAuthors([...authors, { name: "", contactInfo: [] }])}
          >
            Add author
          </button>
        </div>
      )}
    </Modal>
  );
}
