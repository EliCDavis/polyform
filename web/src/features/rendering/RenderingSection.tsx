import { useEffect, useState } from "react";
import type { Color } from "three";
import { useEditorOptional } from "../editor/EditorContext";
import { RenderingOption, RenderingColorOption } from "./RenderingControls";
import { SSAOGroup } from "./SSAOGroup";
import { BloomGroup } from "./BloomGroup";

interface RenderingSectionProps {}

export function RenderingSection() {
  const editor = useEditorOptional();

  const [fov, setFov] = useState<number>(editor.threeApp.Camera.fov);
  const [floorColor, setFloorColor] = useState<string>(
    `#${editor.threeApp.Ground.Material.color.getHexString()}`,
  );
  const [lightColor, setLightColor] = useState<string>(
    `#${editor.threeApp.Lighting.DirLight.color.getHexString()}`,
  );
  const [skyColor, setSkyColor] = useState<string>(
    `#${(editor.threeApp.Scene.background as Color).getHexString()}`,
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
    (editor.threeApp.Scene.background as Color).set(skyColor);
    editor.threeApp.Fog.color.set(skyColor);
  }, [skyColor]);

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
        <RenderingColorOption
          name="sky color"
          description="The color of the background/sky"
          setValue={setSkyColor}
          value={skyColor}
        />
        <SSAOGroup threeApp={editor.threeApp} />
        <BloomGroup threeApp={editor.threeApp} />
      </div>
    </>
  );
}
