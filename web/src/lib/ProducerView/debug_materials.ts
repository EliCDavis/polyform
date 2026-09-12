import {
    CanvasTexture,
    Material,
    MeshBasicMaterial,
    MeshNormalMaterial,
    RepeatWrapping,
    SRGBColorSpace,
} from "three";

export enum ShadingMode {
    Shaded = "shaded",
    Normals = "normals",
    UVs = "uvs",
}

const CHECKER_CELLS = 16;
const CHECKER_CELL_PIXELS = 32;

/**
 * Hue runs along U and lightness along V so a mirrored or rotated island reads
 * differently from a correct one, which a plain two-tone checker can't show.
 */
function createUVCheckerTexture(): CanvasTexture {
    const size = CHECKER_CELLS * CHECKER_CELL_PIXELS;
    const canvas = document.createElement("canvas");
    canvas.width = size;
    canvas.height = size;

    const ctx = canvas.getContext("2d");
    for (let row = 0; row < CHECKER_CELLS; row++) {
        for (let column = 0; column < CHECKER_CELLS; column++) {
            const dark = (row + column) % 2 === 0;
            if (dark) {
                const hue = (column / CHECKER_CELLS) * 360;
                const lightness = 30 + (row / CHECKER_CELLS) * 35;
                ctx.fillStyle = `hsl(${hue}, 65%, ${lightness}%)`;
            } else {
                ctx.fillStyle = row % 2 === 0 ? "#f2f2f2" : "#cfcfcf";
            }
            ctx.fillRect(
                column * CHECKER_CELL_PIXELS,
                // Canvas Y grows downward while V grows upward; flipY on the
                // texture already corrects this, so draw rows in order.
                row * CHECKER_CELL_PIXELS,
                CHECKER_CELL_PIXELS,
                CHECKER_CELL_PIXELS
            );
        }
    }

    ctx.strokeStyle = "rgba(0, 0, 0, 0.55)";
    ctx.lineWidth = 2;
    ctx.strokeRect(1, 1, size - 2, size - 2);

    const texture = new CanvasTexture(canvas);
    texture.wrapS = RepeatWrapping;
    texture.wrapT = RepeatWrapping;
    texture.colorSpace = SRGBColorSpace;
    return texture;
}

let uvMaterial: MeshBasicMaterial | null = null;
let normalMaterial: MeshNormalMaterial | null = null;

export function debugMaterialFor(mode: ShadingMode): Material | null {
    switch (mode) {
        case ShadingMode.Normals:
            if (!normalMaterial) {
                normalMaterial = new MeshNormalMaterial();
            }
            return normalMaterial;

        case ShadingMode.UVs:
            if (!uvMaterial) {
                uvMaterial = new MeshBasicMaterial({ map: createUVCheckerTexture() });
            }
            return uvMaterial;

        default:
            return null;
    }
}
