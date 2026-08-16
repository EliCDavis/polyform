import { useEffect, useState } from "react";
import {
  ACESFilmicToneMapping,
  AgXToneMapping,
  CineonToneMapping,
  LinearToneMapping,
  NeutralToneMapping,
  NoToneMapping,
  ReinhardToneMapping,
  type Color,
  type ToneMapping,
} from "three";
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

const TONE_MAPPING_TO_THREE: Record<ToneMappingOption, ToneMapping> = {
  [ToneMappingOption.None]: NoToneMapping,
  [ToneMappingOption.Linear]: LinearToneMapping,
  [ToneMappingOption.Reinhard]: ReinhardToneMapping,
  [ToneMappingOption.Cineon]: CineonToneMapping,
  [ToneMappingOption.AcesFilmic]: ACESFilmicToneMapping,
  [ToneMappingOption.AgX]: AgXToneMapping,
  [ToneMappingOption.Neutral]: NeutralToneMapping,
};

const THREE_TO_TONE_MAPPING: Record<number, ToneMappingOption> = {
  [NoToneMapping]: ToneMappingOption.None,
  [LinearToneMapping]: ToneMappingOption.Linear,
  [ReinhardToneMapping]: ToneMappingOption.Reinhard,
  [CineonToneMapping]: ToneMappingOption.Cineon,
  [ACESFilmicToneMapping]: ToneMappingOption.AcesFilmic,
  [AgXToneMapping]: ToneMappingOption.AgX,
  [NeutralToneMapping]: ToneMappingOption.Neutral,
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
  const [toneMapping, setToneMapping] = useState<string>(
    THREE_TO_TONE_MAPPING[editor.threeApp.Renderer.toneMapping] ??
      ToneMappingOption.AcesFilmic,
  );
  const [exposure, setExposure] = useState<number>(
    editor.threeApp.Renderer.toneMappingExposure,
  );

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
    editor.threeApp.Renderer.toneMapping = TONE_MAPPING_TO_THREE[toneMapping];
  }, [toneMapping]);

  useEffect(() => {
    editor.threeApp.Renderer.toneMappingExposure = exposure;
  }, [exposure]);

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
        <RenderingOption
          name="exposure"
          description="Overall brightness applied after color grading"
          setValue={setExposure}
          value={exposure}
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
