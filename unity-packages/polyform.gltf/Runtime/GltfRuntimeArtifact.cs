using EliCDavis.Polyform.Artifacts;
using UnityEngine;

namespace EliCDavis.Polyform.GLTF
{
    public class GltfRuntimeArtifact : IRuntimeArtifact
    {
        private GameObject go;

        public GltfRuntimeArtifact(GameObject go)
        {
            this.go = go;
        }

        public void Unload()
        {
            GameObject.Destroy(go);
        }
    }
}