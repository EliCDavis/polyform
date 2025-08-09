using EliCDavis.Polyform.Loading;
using EliCDavis.Polyform.Models;
using UnityEngine;

namespace EliCDavis.Polyform.GLTF
{
    [CreateAssetMenu(fileName = "GLTF Manifest Handler", menuName = "Polyform/Manifest Handlers/Runtime/GLTF", order = 1)]
    public class GltfRuntimeManifestHandler : RuntimeManifestHandler
    {
        public override bool CanHandle(Manifest manifest)
        {
            return manifest.Main.EndsWith(".gltf") || manifest.Main.EndsWith(".glb");
        }

        public override IRuntimeManifestInstance Handle(GameObject parent, Graph graph, ManifestInstance manifestInstance)
        {
            var obj = new GameObject("GLTF Manifest");
            obj.transform.SetParent(parent.transform);
            obj.transform.localPosition = Vector3.zero;
            obj.transform.localRotation = Quaternion.identity;
            
            var gltf = obj.AddComponent<GLTFast.GltfAsset>();
            var url = graph.FormatURl($"manifest/{manifestInstance.Id}/{manifestInstance.Manifest.Main}");
            gltf.Url = url;
            return new GltfRuntimeManifestInstance(obj);
        }
    }
}