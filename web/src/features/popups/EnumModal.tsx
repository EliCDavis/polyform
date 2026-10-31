import { useState } from "react";
import { ChevronDown, ChevronUp, X } from "lucide-react";
import { Modal, PopupButton } from "@/components/Modal";
import type { SchemaManager } from "@/lib/schema_manager";
import type { NodeManager } from "@/lib/node_manager";
import type { EnumBody, EnumDefinition, EnumResponse } from "@/types/schema";
import { getApiErrorMessage, requestManager } from "@/api/client";

interface MemberDraft {
  key: number;
  originalName?: string;
  name: string;
  description: string;
}

interface EnumModalProps {
  open: boolean;
  onClose: () => void;
  schemaManager: SchemaManager;
  nodeManager: NodeManager;
  enumId?: string;
  initial?: EnumDefinition;
}

let nextMemberKey = 0;

function draftMember(name = "", description = "", originalName?: string): MemberDraft {
  return { key: nextMemberKey++, originalName, name, description };
}

function initialMembers(initial?: EnumDefinition): MemberDraft[] {
  if (!initial) return [draftMember()];
  return initial.members.map((member) => draftMember(member.name, member.description ?? "", member.name));
}

export function EnumModal({ open, onClose, schemaManager, nodeManager, enumId, initial }: EnumModalProps) {
  const [name, setName] = useState(initial?.name ?? "New Enum");
  const [description, setDescription] = useState(initial?.description ?? "");
  const [members, setMembers] = useState<MemberDraft[]>(() => initialMembers(initial));

  const trimmedNames = members.map((member) => member.name.trim());
  const duplicate = trimmedNames.some((memberName, i) => memberName && trimmedNames.indexOf(memberName) !== i);
  const valid =
    name.trim().length > 0 &&
    members.length > 0 &&
    trimmedNames.every((memberName) => memberName.length > 0) &&
    !duplicate;

  const updateMember = (key: number, patch: Partial<MemberDraft>) => {
    setMembers((current) => current.map((member) => (member.key === key ? { ...member, ...patch } : member)));
  };

  const removeMember = (key: number) => {
    setMembers((current) => current.filter((member) => member.key !== key));
  };

  const moveMember = (index: number, delta: number) => {
    setMembers((current) => {
      const target = index + delta;
      if (target < 0 || target >= current.length) return current;
      const next = current.slice();
      [next[index], next[target]] = [next[target], next[index]];
      return next;
    });
  };

  const applied = (resp: EnumResponse) => {
    nodeManager.registerCustomNodeType(resp.nodeType);
    nodeManager.registerCustomNodeType(resp.membersNodeType);
    nodeManager.refreshRegisteredTypes(() => {
      if (enumId) {
        // Canvas nodes bake the member list into their widgets; rebuild them.
        nodeManager.clearCanvasNodes();
      }
      schemaManager.refreshSchema(enumId ? "updated an enum" : "created an enum");
    });
    onClose();
  };

  const save = () => {
    if (!valid) return;
    const body: EnumBody = {
      name: name.trim(),
      description: description.trim(),
      members: members.map((member) => ({
        name: member.name.trim(),
        description: member.description.trim() || undefined,
      })),
    };

    if (!enumId) {
      requestManager.createEnum(body, applied, (err) => alert(getApiErrorMessage(err, "Failed to create enum")));
      return;
    }

    const renamed: { [oldName: string]: string } = {};
    for (const member of members) {
      if (member.originalName && member.originalName !== member.name.trim()) {
        renamed[member.originalName] = member.name.trim();
      }
    }
    if (Object.keys(renamed).length > 0) {
      body.renamed = renamed;
    }
    requestManager.updateEnum(enumId, body, applied, (err) => alert(getApiErrorMessage(err, "Failed to update enum")));
  };

  return (
    <Modal
      title={enumId ? "Edit Enum" : "New Enum"}
      open={open}
      onClose={onClose}
      actions={
        <>
          <PopupButton onClick={onClose}>Cancel</PopupButton>
          <PopupButton variant="primary" onClick={save} disabled={!valid}>
            {enumId ? "Save" : "Create"}
          </PopupButton>
        </>
      }
    >
      <div style={{ display: "flex", flexDirection: "column", gap: 12, width: 480 }}>
        <label>Name</label>
        <input type="text" value={name} onChange={(e) => setName(e.target.value)} />
        <label>Description</label>
        <textarea value={description} onChange={(e) => setDescription(e.target.value)} />
        <label>Members</label>
        <div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
          {members.map((member, index) => (
            <div key={member.key} style={{ display: "flex", flexDirection: "row", alignItems: "center", gap: 6 }}>
              <input
                type="text"
                placeholder="Name"
                value={member.name}
                onChange={(e) => updateMember(member.key, { name: e.target.value })}
              />
              <input
                type="text"
                placeholder="Description"
                value={member.description}
                onChange={(e) => updateMember(member.key, { description: e.target.value })}
              />
              <button
                type="button"
                className="icon-button"
                aria-label="Move up"
                disabled={index === 0}
                onClick={() => moveMember(index, -1)}
              >
                <ChevronUp size={14} />
              </button>
              <button
                type="button"
                className="icon-button"
                aria-label="Move down"
                disabled={index === members.length - 1}
                onClick={() => moveMember(index, 1)}
              >
                <ChevronDown size={14} />
              </button>
              <button
                type="button"
                className="icon-button"
                aria-label="Remove member"
                onClick={() => removeMember(member.key)}
              >
                <X size={14} />
              </button>
            </div>
          ))}
        </div>
        {duplicate && <span className="field-hint">Member names must be unique.</span>}
        <button type="button" onClick={() => setMembers((current) => [...current, draftMember()])}>
          Add Member
        </button>
      </div>
    </Modal>
  );
}
