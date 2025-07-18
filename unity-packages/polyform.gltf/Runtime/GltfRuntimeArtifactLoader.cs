using EliCDavis.Polyform.Artifacts;
using EliCDavis.Polyform.Models;
using UnityEngine;

namespace EliCDavis.Polyform.GLTF
{
    [CreateAssetMenu(fileName = "GLTF Artifact Handler", menuName = "Polyform/Artifact Handlers/Runtime/GLTF", order = 1)]
    public class GltfRuntimeArtifactLoader : RuntimeArtifactLoader
    {
        public override bool CanHandle(Manifest manifest)
        {
            return manifest.Main.EndsWith(".gltf") || manifest.Main.EndsWith(".glb");
        }

        public override IRuntimeArtifact Handle(GameObject parent, Graph graph, ManifestInstance manifestInstance)
        {
            var obj = new GameObject("GLTF Artifact");
            obj.transform.SetParent(parent.transform);
            obj.transform.localPosition = Vector3.zero;
            obj.transform.localRotation = Quaternion.identity;
            
            var gltf = obj.AddComponent<GLTFast.GltfAsset>();
            var url = graph.FormatURl($"manifest/{manifestInstance.Id}/{manifestInstance.Manifest.Main}");
            gltf.Url = url;
            return new GltfRuntimeArtifact(obj);
        }
    }
}