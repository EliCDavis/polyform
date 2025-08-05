using EliCDavis.Polyform.Loading;
using UnityEngine;

namespace EliCDavis.Polyform.GLTF
{
    public class GltfRuntimeManifest : IRuntimeManifest
    {
        private GameObject go;

        public GltfRuntimeManifest(GameObject go)
        {
            this.go = go;
        }

        public void Unload()
        {
            GameObject.Destroy(go);
        }
    }
}