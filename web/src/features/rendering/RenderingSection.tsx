import { useEffect, useRef, useState } from "react";
import { ShadingMode } from "@/lib/ProducerView/debug_materials";
import { BlendFunction, ToneMappingMode } from "postprocessing";
import type { Color } from "three";
import { useEditorOptional } from "../editor/EditorContext";
import {
  RenderingOption,
  RenderingColorOption,
  RenderingSelectOption,
  RenderingSliderOption,
  RenderingToggleOption,
  RenderingGroup,
} from "./RenderingControls";
import { SSAOGroup } from "./SSAOGroup";
import { BloomGroup } from "./BloomGroup";
import { AutoRotateGroup } from "./AutoRotateGroup";

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

const SHADING_OPTIONS: Array<{ label: string; value: ShadingMode }> = [
  { label: "Shaded", value: ShadingMode.Shaded },
  { label: "Normals", value: ShadingMode.Normals },
  { label: "UV Checker", value: ShadingMode.UVs },
];

const TONE_MAPPING_OPTIONS: Array<{ label: string; value: string }> = [
  { label: "None", value: ToneMappingOption.None },
  { label: "Linear", value: ToneMappingOption.Linear },
  { label: "Reinhard", value: ToneMappingOption.Reinhard },
  { label: "Cineon", value: ToneMappingOption.Cineon },
  { label: "ACES Filmic", value: ToneMappingOption.AcesFilmic },
  { label: "AgX", value: ToneMappingOption.AgX },
  { label: "Neutral", value: ToneMappingOption.Neutral },
];

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
  const [wireframe, setWireframe] = useState<boolean>(
    editor.producerViewManager.wireframe,
  );
  const [shading, setShading] = useState<ShadingMode>(
    editor.producerViewManager.shading,
  );
  const [showFloor, setShowFloor] = useState<boolean>(
    editor.threeApp.Ground.Mesh.visible,
  );
  const [showGrid, setShowGrid] = useState<boolean>(editor.threeApp.Grid.visible);
  const [shadows, setShadows] = useState<boolean>(
    editor.threeApp.Lighting.DirLight.castShadow,
  );
  const [lightAzimuth, setLightAzimuth] = useState<number>(() => {
    const p = editor.threeApp.Lighting.DirLight.position;
    return Math.round((Math.atan2(p.x, p.z) * 180) / Math.PI);
  });
  const [lightElevation, setLightElevation] = useState<number>(() => {
    const p = editor.threeApp.Lighting.DirLight.position;
    return Math.round((Math.asin(p.y / Math.max(p.length(), 1e-6)) * 180) / Math.PI);
  });
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
    editor.producerViewManager.SetWireframe(wireframe);
  }, [wireframe]);

  useEffect(() => {
    editor.producerViewManager.SetShading(shading);
  }, [shading]);

  useEffect(() => {
    editor.threeApp.Ground.Mesh.visible = showFloor;
  }, [showFloor]);

  useEffect(() => {
    editor.threeApp.Grid.visible = showGrid;
  }, [showGrid]);

  useEffect(() => {
    editor.threeApp.Lighting.DirLight.castShadow = shadows;
  }, [shadows]);

  useEffect(() => {
    const light = editor.threeApp.Lighting.DirLight;
    const distance = Math.max(light.position.length(), 1e-6);
    const azimuth = (lightAzimuth * Math.PI) / 180;
    const elevation = (lightElevation * Math.PI) / 180;
    const horizontal = Math.cos(elevation) * distance;
    light.position.set(
      Math.sin(azimuth) * horizontal,
      Math.sin(elevation) * distance,
      Math.cos(azimuth) * horizontal,
    );
    editor.producerViewManager.RefitShadowCamera();
  }, [lightAzimuth, lightElevation]);

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
          name="sky color"
          description="The color of the background/sky"
          setValue={setSkyColor}
          value={skyColor}
        />
        <RenderingToggleOption
          name="show floor"
          description="Draws the ground plane beneath the model"
          setValue={setShowFloor}
          value={showFloor}
        />
        <RenderingToggleOption
          name="show grid"
          description="Draws a measuring grid on the ground, sized to the model"
          setValue={setShowGrid}
          value={showGrid}
        />
        <RenderingSelectOption
          name="color grading"
          description="The tone mapping algorithm used to map the scene's lighting to displayable colors"
          setValue={setToneMapping}
          value={toneMapping}
          options={TONE_MAPPING_OPTIONS}
        />
        <RenderingSelectOption
          name="shading"
          description="Replaces the model's materials with a debug visualization"
          setValue={setShading}
          value={shading}
          options={SHADING_OPTIONS}
        />
        <RenderingToggleOption
          name="wireframe"
          description="Draws the model's edges instead of its shaded surfaces"
          setValue={setWireframe}
          value={wireframe}
        />
        <RenderingGroup name="light">
          <RenderingColorOption
            name="color"
            description="The color of the scene's lighting"
            setValue={setLightColor}
            value={lightColor}
          />
          <RenderingOption
            name="intensity"
            description="How strongly the scene's lighting contributes to the render"
            setValue={setLightIntensity}
            value={lightIntensity}
          />
          <RenderingSliderOption
            name="angle"
            description="Compass direction the light comes from"
            setValue={setLightAzimuth}
            value={lightAzimuth}
            min={-180}
            max={180}
            unit="°"
          />
          <RenderingSliderOption
            name="elevation"
            description="Height of the light above the horizon"
            setValue={setLightElevation}
            value={lightElevation}
            min={-90}
            max={90}
            unit="°"
          />
          <RenderingToggleOption
            name="shadows"
            description="Casts shadows from the scene's directional light"
            setValue={setShadows}
            value={shadows}
          />
        </RenderingGroup>
        <SSAOGroup
          threeApp={editor.threeApp}
          producerViewManager={editor.producerViewManager}
        />
        <BloomGroup threeApp={editor.threeApp} />
        <AutoRotateGroup threeApp={editor.threeApp} />
      </div>
    </>
  );
}
