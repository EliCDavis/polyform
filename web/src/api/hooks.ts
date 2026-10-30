import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useRef } from "react";
import {
  fetchSchema,
  fetchNodeTypes,
  fetchStartedTime,
} from "./client";

export const schemaQueryKey = ["schema"] as const;
export const nodeTypesQueryKey = ["nodeTypes"] as const;
export const startedQueryKey = ["started"] as const;

export function useSchema() {
  return useQuery({
    queryKey: schemaQueryKey,
    queryFn: fetchSchema,
  });
}

export function useNodeTypes() {
  return useQuery({
    queryKey: nodeTypesQueryKey,
    queryFn: fetchNodeTypes,
    staleTime: Infinity,
  });
}

export function useStartedPolling(onModelVersion: (version: number) => void) {
  const initTimeRef = useRef<string | null>(null);

  return useQuery({
    queryKey: startedQueryKey,
    queryFn: async () => {
      const payload = await fetchStartedTime();
      if (initTimeRef.current === null) {
        initTimeRef.current = payload.time;
      } else if (initTimeRef.current !== payload.time) {
        location.reload();
      }
      onModelVersion(payload.modelVersion);
      return payload;
    },
    refetchInterval: 1000,
  });
}

export function useInvalidateSchema() {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ queryKey: schemaQueryKey });
}
