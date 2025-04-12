import GUI from "three/examples/jsm/libs/lil-gui.module.min";
import { ViewportManager, ViewportSetting } from "../viewport_manager";
import { ViewportSettings } from "../viewport_settings";
import { Color } from "three";
import { ThreeApp } from "../three_app";
import { ProducerViewManager } from "../ProducerView/producer_view_manager";

export function BuildRenderingSetting(
    parentPanel: GUI,
    viewportManager: ViewportManager,
    viewportSettings: ViewportSettings,
    threeApp: ThreeApp,
    producerViewManager: ProducerViewManager
) {
    viewportManager.AddSetting(
        "renderWireframe",
        new ViewportSetting(
            "renderWireframe",
            viewportSettings,
            parentPanel
                .add(viewportSettings, "renderWireframe")
                .name("Render Wireframe"),
            () => {
                producerViewManager.SetWireframe(viewportSettings.renderWireframe);
            }
        )
    )

    viewportManager.AddSetting(
        "background",
        new ViewportSetting(
            "background",
            viewportSettings,
            parentPanel
                .addColor(viewportSettings, "background")
                .name("Background"),
            () => {
                threeApp.Scene.background = new Color(viewportSettings.background);
            }
        )
    );

    viewportManager.AddSetting(
        "lighting",
        new ViewportSetting(
            "lighting",
            viewportSettings,
            parentPanel
                .addColor(viewportSettings, "lighting")
                .name("Lighting"),
            () => {
                threeApp.Lighting.DirLight.color = new Color(viewportSettings.lighting);
                threeApp.Lighting.HemiLight.color = new Color(viewportSettings.lighting);
            },
        )
    );

    viewportManager.AddSetting(
        "ground",
        new ViewportSetting(
            "ground",
            viewportSettings,
            parentPanel
                .addColor(viewportSettings, "ground")
                .name("Ground"),
            () => {
                threeApp.Ground.Material.color = new Color(viewportSettings.ground);
            }
        )
    );


}

export function BuildFogSettings(
    parentPanel: GUI,
    viewportManager: ViewportManager,
    viewportSettings: ViewportSettings,
    threeApp: ThreeApp
) {
    const fogSettingsFolder = parentPanel.addFolder("Fog");
    fogSettingsFolder.close();

    viewportManager.AddSetting(
        "fog/color",
        new ViewportSetting(
            "color",
            viewportSettings.fog,
            fogSettingsFolder.addColor(viewportSettings.fog, "color"),
            () => {
                threeApp.Fog.color = new Color(viewportSettings.fog.color);
            }
        )
    );

    viewportManager.AddSetting(
        "fog/near",
        new ViewportSetting(
            "near",
            viewportSettings.fog,
            fogSettingsFolder.add(viewportSettings.fog, "near"),
            () => {
                threeApp.Fog.near = viewportSettings.fog.near;
            }
        )
    );

    viewportManager.AddSetting(
        "fog/far",
        new ViewportSetting(
            "far",
            viewportSettings.fog,
            fogSettingsFolder.add(viewportSettings.fog, "far"),
            () => {
                threeApp.Fog.far = viewportSettings.fog.far;
            }
        )
    );
}