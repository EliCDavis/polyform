import { useEffect, useRef, useState } from "react";
import { BlendFunction } from "postprocessing";
import type { ThreeApp } from "@/lib/three_app";
import { RenderingGroup, RenderingOption } from "./RenderingControls";

interface BloomGroupProps {
  threeApp: ThreeApp;
}

export function BloomGroup({ threeApp }: BloomGroupProps) {
  const bloom = threeApp.PostProcessing.Bloom;
  const enabledBlendFunction = useRef(bloom.blendMode.blendFunction);

  const [enabled, setEnabled] = useState<boolean>(
    bloom.blendMode.blendFunction !== BlendFunction.SKIP,
  );
  const [strength, setStrength] = useState<number>(bloom.intensity);
  const [radius, setRadius] = useState<number>(bloom.mipmapBlurPass.radius);
  const [threshold, setThreshold] = useState<number>(
    bloom.luminanceMaterial.threshold,
  );

  useEffect(() => {
    bloom.blendMode.blendFunction = enabled
      ? enabledBlendFunction.current
      : BlendFunction.SKIP;
  }, [enabled]);

  useEffect(() => {
    bloom.intensity = strength;
  }, [strength]);

  useEffect(() => {
    bloom.mipmapBlurPass.radius = radius;
  }, [radius]);

  useEffect(() => {
    bloom.luminanceMaterial.threshold = threshold;
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
