import { useEffect, useRef, useState } from "react";
import { BlendFunction, ToneMappingMode } from "postprocessing";
import type { Color } from "three";
import { useEditorOptional } from "../editor/EditorContext";
import {
  RenderingOption,
  RenderingColorOption,
  RenderingSelectOption,
} from "./RenderingControls";
import { SSAOGroup } from "./SSAOGroup";
import { BloomGroup } from "./BloomGroup";

interface RenderingSectionProps {}

enum ToneMappingOption {
  None = "none",
  Linear = "linear",
  Reinhard = "reinhard",
  Cineon = "cineon",
  AcesFilmic = "aces-filmic",
  AgX = "agx",
  Neutral = "neutral",
}

const TONE_MAPPING_OPTIONS: Array<{ label: string; value: string }> = [
  { label: "None", value: ToneMappingOption.None },
  { label: "Linear", value: ToneMappingOption.Linear },
  { label: "Reinhard", value: ToneMappingOption.Reinhard },
  { label: "Cineon", value: ToneMappingOption.Cineon },
  { label: "ACES Filmic", value: ToneMappingOption.AcesFilmic },
  { label: "AgX", value: ToneMappingOption.AgX },
  { label: "Neutral", value: ToneMappingOption.Neutral },
];

// "None" isn't a real ToneMappingMode — it's represented by skipping the
// effect entirely (see the blendFunction handling below), so this map only
// needs to cover the modes that actually get applied.
const TONE_MAPPING_TO_MODE: Partial<Record<ToneMappingOption, ToneMappingMode>> = {
  [ToneMappingOption.Linear]: ToneMappingMode.LINEAR,
  [ToneMappingOption.Reinhard]: ToneMappingMode.REINHARD,
  [ToneMappingOption.Cineon]: ToneMappingMode.CINEON,
  [ToneMappingOption.AcesFilmic]: ToneMappingMode.ACES_FILMIC,
  [ToneMappingOption.AgX]: ToneMappingMode.AGX,
  [ToneMappingOption.Neutral]: ToneMappingMode.NEUTRAL,
};

const MODE_TO_TONE_MAPPING: Partial<Record<ToneMappingMode, ToneMappingOption>> = {
  [ToneMappingMode.LINEAR]: ToneMappingOption.Linear,
  [ToneMappingMode.REINHARD]: ToneMappingOption.Reinhard,
  [ToneMappingMode.CINEON]: ToneMappingOption.Cineon,
  [ToneMappingMode.ACES_FILMIC]: ToneMappingOption.AcesFilmic,
  [ToneMappingMode.AGX]: ToneMappingOption.AgX,
  [ToneMappingMode.NEUTRAL]: ToneMappingOption.Neutral,
};

export function RenderingSection() {
  const editor = useEditorOptional();

  const [fov, setFov] = useState<number>(editor.threeApp.Camera.fov);
  const [floorColor, setFloorColor] = useState<string>(
    `#${editor.threeApp.Ground.Material.color.getHexString()}`,
  );
  const [lightColor, setLightColor] = useState<string>(
    `#${editor.threeApp.Lighting.DirLight.color.getHexString()}`,
  );
  const [lightIntensity, setLightIntensity] = useState<number>(
    editor.threeApp.Lighting.DirLight.intensity,
  );
  const [skyColor, setSkyColor] = useState<string>(
    `#${(editor.threeApp.Scene.background as Color).getHexString()}`,
  );
  const toneMappingEffect = editor.threeApp.PostProcessing.ToneMapping;
  // "None" is represented by skipping the effect (BlendFunction.SKIP)
  // rather than a real mode, so remember the last real blend function to
  // restore when switching back to any actual tone-mapping mode.
  const enabledBlendFunction = useRef(
    toneMappingEffect.blendMode.blendFunction === BlendFunction.SKIP
      ? BlendFunction.SRC
      : toneMappingEffect.blendMode.blendFunction,
  );
  const [toneMapping, setToneMapping] = useState<string>(() => {
    if (toneMappingEffect.blendMode.blendFunction === BlendFunction.SKIP) {
      return ToneMappingOption.None;
    }
    return MODE_TO_TONE_MAPPING[toneMappingEffect.mode] ?? ToneMappingOption.AcesFilmic;
  });

  useEffect(() => {
    editor.threeApp.Camera.fov = fov;
  }, [fov]);

  useEffect(() => {
    editor.threeApp.Ground.Material.color.set(floorColor);
  }, [floorColor]);

  useEffect(() => {
    editor.threeApp.Lighting.DirLight.color.set(lightColor);
    editor.threeApp.Lighting.HemiLight.color.set(lightColor);
  }, [lightColor]);

  useEffect(() => {
    editor.threeApp.Lighting.DirLight.intensity = lightIntensity;
    editor.threeApp.Lighting.HemiLight.intensity = lightIntensity;
  }, [lightIntensity]);

  useEffect(() => {
    (editor.threeApp.Scene.background as Color).set(skyColor);
    editor.threeApp.Fog.color.set(skyColor);
  }, [skyColor]);

  useEffect(() => {
    if (toneMapping === ToneMappingOption.None) {
      toneMappingEffect.blendMode.blendFunction = BlendFunction.SKIP;
      return;
    }
    toneMappingEffect.mode = TONE_MAPPING_TO_MODE[toneMapping as ToneMappingOption]!;
    toneMappingEffect.blendMode.blendFunction = enabledBlendFunction.current;
  }, [toneMapping]);

  return (
    <>
      <div className="sidebar-header">Rendering</div>
      <div className="sidebar-section-content">
        <RenderingOption
          name="fov"
          description="The vertical field of view, from bottom to top of view, in degrees"
          setValue={setFov}
          value={fov}
        />
        <RenderingColorOption
          name="floor color"
          description="The color of the ground plane"
          setValue={setFloorColor}
          value={floorColor}
        />
        <RenderingColorOption
          name="light color"
          description="The color of the scene's lighting"
          setValue={setLightColor}
          value={lightColor}
        />
        <RenderingOption
          name="light intensity"
          description="How strongly the scene's lighting contributes to the render"
          setValue={setLightIntensity}
          value={lightIntensity}
        />
        <RenderingColorOption
          name="sky color"
          description="The color of the background/sky"
          setValue={setSkyColor}
          value={skyColor}
        />
        <RenderingSelectOption
          name="color grading"
          description="The tone mapping algorithm used to map the scene's lighting to displayable colors"
          setValue={setToneMapping}
          value={toneMapping}
          options={TONE_MAPPING_OPTIONS}
        />
        <SSAOGroup
          threeApp={editor.threeApp}
          producerViewManager={editor.producerViewManager}
        />
        <BloomGroup threeApp={editor.threeApp} />
      </div>
    </>
  );
}
