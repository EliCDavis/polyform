using EliCDavis.Polyform.Loading;
using UnityEngine;

namespace EliCDavis.Polyform.GLTF
{
    public class GltfRuntimeManifestInstance : IRuntimeManifestInstance
    {
        private GameObject go;

        public GltfRuntimeManifestInstance(GameObject go)
        {
            this.go = go;
        }

        public void Unload()
        {
            GameObject.Destroy(go);
        }
    }
}