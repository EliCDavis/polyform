using System;
using System.Collections;
using EliCDavis.Polyform.Artifacts;
using UnityEngine;

namespace EliCDavis.Polyform
{
    public class PolyformRenderer : MonoBehaviour
    {
        [SerializeField] private ConnectionConfig connectionConfig;

        [SerializeField] private ArtifactHandler[] handlers;

        private void Start()
        {
            Render();
        }

        public void Render()
        {
            StartCoroutine(Run());
        }

        private IEnumerator Run()
        {
            var manifestsReq = connectionConfig.AvailableManifests();
            yield return manifestsReq.Run();

            foreach (var manifest in manifestsReq.Result)
            {
                yield return LoadManifest(manifest);
            }
        }

        private IEnumerator LoadManifest(AvailableManifest manifest)
        {
            Debug.Log($"Loading: {manifest.Name}/{manifest.Port}");
            var manifestsReq = connectionConfig.CreateManifest(manifest.Name, manifest.Port);
            yield return manifestsReq.Run();

            var gltf = gameObject.AddComponent<GLTFast.GltfAsset>();
            gltf.Url = $"http://localhost:8080/manifest/{manifestsReq.Result.Id}/{manifestsReq.Result.Manifest.Main}";
        }
    }
}