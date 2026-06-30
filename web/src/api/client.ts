export {
  RequestManager,
  downloadBlob,
  saveFileToDisk,
  type StartedResponse,
  type CreateNodeResponse,
  type SetProducerBody,
} from "../lib/requests";

import { RequestManager } from "../lib/requests";
import type { GraphInstance, GraphExecutionReport, RegisteredTypes } from "../types/schema";

export const requestManager = new RequestManager();

export function fetchSchema(): Promise<GraphInstance> {
  return new Promise((resolve) => {
    requestManager.getSchema(resolve);
  });
}

export function fetchNodeTypes(): Promise<RegisteredTypes> {
  return new Promise((resolve) => {
    requestManager.getNodeTypes(resolve);
  });
}

export function fetchStartedTime(): Promise<{ time: string; modelVersion: number }> {
  return new Promise((resolve) => {
    requestManager.getStartedTime(resolve);
  });
}

export function fetchExecutionReport(): Promise<GraphExecutionReport> {
  return new Promise((resolve) => {
    requestManager.getExecutionReport(resolve);
  });
}

export function fetchGraph(): Promise<GraphInstance> {
  return new Promise((resolve) => {
    requestManager.getGraph(resolve);
  });
}

export function getApiErrorMessage(err: unknown, fallback: string): string {
  if (typeof err === "object" && err !== null && "error" in err) {
    const message = (err as { error?: unknown }).error;
    if (typeof message === "string" && message) return message;
  }
  return fallback;
}
