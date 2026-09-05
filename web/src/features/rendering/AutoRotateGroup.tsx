import { useEffect, useState } from "react";
import type { ThreeApp } from "@/lib/three_app";
import { RenderingGroup, RenderingOption } from "./RenderingControls";

interface AutoRotateGroupProps {
  threeApp: ThreeApp;
}

export function AutoRotateGroup({ threeApp }: AutoRotateGroupProps) {
  const controls = threeApp.OrbitControls;

  const [enabled, setEnabled] = useState<boolean>(controls.autoRotate);
  const [speed, setSpeed] = useState<number>(controls.autoRotateSpeed);

  useEffect(() => {
    controls.autoRotate = enabled;
  }, [enabled]);

  useEffect(() => {
    controls.autoRotateSpeed = speed;
  }, [speed]);

  return (
    <RenderingGroup
      name="auto rotate"
      description="Continuously orbits the camera around the target"
      enabled={enabled}
      setEnabled={setEnabled}
    >
      <RenderingOption
        name="speed"
        description="How fast the camera orbits around the target"
        setValue={setSpeed}
        value={speed}
      />
    </RenderingGroup>
  );
}
