import { requestManager } from "./client";

export function setVariableValue(key: string, value: unknown): Promise<Response> {
  return fetch(`./variable/value/${key}`, {
    method: "POST",
    body: JSON.stringify(value),
  });
}

export function setBinaryVariableValue(key: string, onSuccess?: () => void): void {
  const input = document.createElement("input");
  input.type = "file";
  input.onchange = (e) => {
    const file = (e.target as HTMLInputElement).files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = (ev) => {
      const content = ev.target?.result;
      requestManager.postBinaryEmptyResponse(
        `./variable/value/${key}`,
        content,
        onSuccess ?? (() => {})
      );
    };
    reader.readAsArrayBuffer(file);
  };
  input.click();
}
