import { useEffect, useState } from "react";
import type { ThreeApp } from "@/lib/three_app";
import { RenderingGroup, RenderingOption } from "./RenderingControls";

interface SSAOGroupProps {
  threeApp: ThreeApp;
}

export function SSAOGroup({ threeApp }: SSAOGroupProps) {
  const ssao = threeApp.PostProcessing.SSAO;

  const [enabled, setEnabled] = useState<boolean>(ssao.enabled);
  const [kernelRadius, setKernelRadius] = useState<number>(ssao.kernelRadius);
  const [minDistance, setMinDistance] = useState<number>(ssao.minDistance);
  const [maxDistance, setMaxDistance] = useState<number>(ssao.maxDistance);

  useEffect(() => {
    ssao.enabled = enabled;
  }, [enabled]);

  useEffect(() => {
    ssao.kernelRadius = kernelRadius;
  }, [kernelRadius]);

  useEffect(() => {
    ssao.minDistance = minDistance;
  }, [minDistance]);

  useEffect(() => {
    ssao.maxDistance = maxDistance;
  }, [maxDistance]);

  return (
    <RenderingGroup
      name="ambient occlusion"
      description="Darkens creases, holes, and surfaces that are close together"
      enabled={enabled}
      setEnabled={setEnabled}
    >
      <RenderingOption
        name="kernel radius"
        description="The sample radius used to estimate occlusion"
        setValue={setKernelRadius}
        value={kernelRadius}
      />
      <RenderingOption
        name="min distance"
        description="Occlusion sample distance where the effect starts fading in"
        setValue={setMinDistance}
        value={minDistance}
      />
      <RenderingOption
        name="max distance"
        description="Occlusion sample distance where the effect fades out"
        setValue={setMaxDistance}
        value={maxDistance}
      />
    </RenderingGroup>
  );
}
