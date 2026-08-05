import { useEffect, useState } from "react";
import type { ThreeApp } from "@/lib/three_app";
import { RenderingGroup, RenderingOption } from "./RenderingControls";

interface BloomGroupProps {
  threeApp: ThreeApp;
}

export function BloomGroup({ threeApp }: BloomGroupProps) {
  const bloom = threeApp.PostProcessing.Bloom;

  const [enabled, setEnabled] = useState<boolean>(bloom.enabled);
  const [strength, setStrength] = useState<number>(bloom.strength);
  const [radius, setRadius] = useState<number>(bloom.radius);
  const [threshold, setThreshold] = useState<number>(bloom.threshold);

  useEffect(() => {
    bloom.enabled = enabled;
  }, [enabled]);

  useEffect(() => {
    bloom.strength = strength;
  }, [strength]);

  useEffect(() => {
    bloom.radius = radius;
  }, [radius]);

  useEffect(() => {
    bloom.threshold = threshold;
  }, [threshold]);

  return (
    <RenderingGroup
      name="bloom"
      description="Makes bright areas of the scene glow"
      enabled={enabled}
      setEnabled={setEnabled}
    >
      <RenderingOption
        name="strength"
        description="The intensity of the glow"
        setValue={setStrength}
        value={strength}
      />
      <RenderingOption
        name="radius"
        description="How far the glow spreads"
        setValue={setRadius}
        value={radius}
      />
      <RenderingOption
        name="threshold"
        description="The minimum brightness a pixel needs to glow"
        setValue={setThreshold}
        value={threshold}
      />
    </RenderingGroup>
  );
}
