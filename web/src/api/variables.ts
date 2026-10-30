import { requestManager } from "./client";
import { variableValueUrl } from "@/lib/requests";

export function setVariableValue(key: string, value: unknown): Promise<Response> {
  return requestManager.setVariableValue(key, value);
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
        variableValueUrl(key),
        content,
        onSuccess ?? (() => {})
      );
    };
    reader.readAsArrayBuffer(file);
  };
  input.click();
}
