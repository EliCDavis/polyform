import { Material, Mesh, Object3D, Texture, WebGLRenderer } from "three";

export function applyMaxAnisotropy(root: Object3D, renderer: WebGLRenderer): void {
  const anisotropy = renderer.capabilities.getMaxAnisotropy();
  root.traverse((object) => {
    if (!(object instanceof Mesh)) return;
    const materials: Material[] = Array.isArray(object.material) ? object.material : [object.material];
    for (const material of materials) {
      for (const value of Object.values(material)) {
        if (value instanceof Texture && value.anisotropy !== anisotropy) {
          value.anisotropy = anisotropy;
          value.needsUpdate = true;
        }
      }
    }
  });
}
