import { useEffect, useRef, useState } from "react";
import { BlendFunction } from "postprocessing";
import type { ThreeApp } from "@/lib/three_app";
import type { ProducerViewManager } from "@/lib/ProducerView/producer_view_manager";
import { RenderingGroup, RenderingOption } from "./RenderingControls";

interface SSAOGroupProps {
  threeApp: ThreeApp;
  producerViewManager: ProducerViewManager;
}

export function SSAOGroup({ threeApp, producerViewManager }: SSAOGroupProps) {
  const ssao = threeApp.PostProcessing.SSAO;
  const ssaoMaterial = ssao.ssaoMaterial;
  // SSAO is merged into a shared EffectPass with Bloom, so there's no
  // per-effect pass to flip `.enabled` on — disable by switching this
  // effect's blend function to SKIP instead, remembering its real one so
  // re-enabling restores it rather than guessing a blend mode.
  const enabledBlendFunction = useRef(ssao.blendMode.blendFunction);

  const [enabled, setEnabled] = useState<boolean>(
    ssao.blendMode.blendFunction !== BlendFunction.SKIP,
  );
  const [radius, setRadius] = useState<number>(ssaoMaterial.radius);
  const [proximityThreshold, setProximityThreshold] = useState<number>(
    ssaoMaterial.worldProximityThreshold,
  );
  const [intensity, setIntensity] = useState<number>(ssao.intensity);

  // radius/proximityThreshold get rescaled to each newly loaded model's
  // size, so pull the sliders back in sync whenever that happens instead
  // of leaving them showing whatever they said on mount.
  useEffect(() => {
    const onRefresh = () => {
      setRadius(ssaoMaterial.radius);
      setProximityThreshold(ssaoMaterial.worldProximityThreshold);
    };
    producerViewManager.SubscribeToCompleteRefresh(onRefresh);
  }, [producerViewManager]);

  useEffect(() => {
    ssao.blendMode.blendFunction = enabled
      ? enabledBlendFunction.current
      : BlendFunction.SKIP;
  }, [enabled]);

  useEffect(() => {
    ssaoMaterial.radius = radius;
  }, [radius]);

  useEffect(() => {
    ssaoMaterial.worldProximityThreshold = proximityThreshold;
  }, [proximityThreshold]);

  useEffect(() => {
    ssao.intensity = intensity;
  }, [intensity]);

  return (
    <RenderingGroup
      name="ambient occlusion"
      description="Darkens creases, holes, and surfaces that are close together"
      enabled={enabled}
      setEnabled={setEnabled}
    >
      <RenderingOption
        name="radius"
        description="The occlusion sampling radius, relative to screen resolution"
        setValue={setRadius}
        value={radius}
      />
      <RenderingOption
        name="proximity threshold"
        description="World-space distance at which two surfaces are considered close enough to occlude each other"
        setValue={setProximityThreshold}
        value={proximityThreshold}
      />
      <RenderingOption
        name="intensity"
        description="The overall strength of the occlusion darkening"
        setValue={setIntensity}
        value={intensity}
      />
    </RenderingGroup>
  );
}
