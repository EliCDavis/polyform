import { useEffect, useState } from "react";
import type { ProducerViewManager } from "@/lib/ProducerView/producer_view_manager";
import { requestRender } from "@/lib/render_scheduler";
import { RenderingGroup, RenderingOption, RenderingToggleOption } from "./RenderingControls";

interface FogGroupProps {
  producerViewManager: ProducerViewManager;
}

export function FogGroup({ producerViewManager }: FogGroupProps) {
  const fog = producerViewManager.fog;

  const [enabled, setEnabled] = useState<boolean>(producerViewManager.fogEnabled);
  const [auto, setAuto] = useState<boolean>(producerViewManager.fogAuto);
  const [near, setNear] = useState<number>(fog.near);
  const [far, setFar] = useState<number>(fog.far);

  useEffect(() => {
    return producerViewManager.SubscribeToCompleteRefresh(() => {
      setNear(fog.near);
      setFar(fog.far);
    });
  }, [producerViewManager]);

  useEffect(() => {
    requestRender();
  });

  useEffect(() => {
    producerViewManager.SetFogEnabled(enabled);
  }, [enabled]);

  useEffect(() => {
    producerViewManager.SetFogAuto(auto);
    if (auto) {
      setNear(fog.near);
      setFar(fog.far);
    }
  }, [auto]);

  useEffect(() => {
    producerViewManager.SetFogRange(near, far);
  }, [near, far]);

  const setManual = (set: (value: number) => void) => (value: number) => {
    setAuto(false);
    set(value);
  };

  return (
    <RenderingGroup
      name="fog"
      description="Fades distant geometry into the sky color"
      enabled={enabled}
      setEnabled={setEnabled}
    >
      <RenderingToggleOption
        name="fit to model"
        description="Sets near and far from the model's size on every load"
        setValue={setAuto}
        value={auto}
      />
      <RenderingOption
        name="near"
        description="Distance from the camera where fog begins"
        setValue={setManual(setNear)}
        value={near}
      />
      <RenderingOption
        name="far"
        description="Distance from the camera where fog is fully opaque"
        setValue={setManual(setFar)}
        value={far}
      />
    </RenderingGroup>
  );
}
