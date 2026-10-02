import { useEffect, useRef, useState } from "react";
import { useInvalidateSchema } from "@/api/hooks";

// Binds a control to a server value: follows it when something else changes
// it, without fighting the user while their own write is still in the air.
// Syncs on the identity of value, so pass a primitive or a stable reference.
export function useLiveValue<T>(
  value: T,
  write: (next: T) => Promise<unknown>,
): [T, (next: T) => void, (next: T) => void] {
  const [shown, setShown] = useState(value);
  const invalidate = useInvalidateSchema();

  const pending = useRef(0);
  const dirty = useRef(false);

  useEffect(() => {
    if (pending.current > 0 || dirty.current) return;
    setShown(value);
  }, [value]);

  const set = (next: T) => {
    dirty.current = true;
    setShown(next);
  };

  const commit = (next: T) => {
    dirty.current = false;
    setShown(next);
    pending.current += 1;
    void write(next)
      .then(() => invalidate())
      .finally(() => {
        pending.current -= 1;
      });
  };

  return [shown, set, commit];
}
