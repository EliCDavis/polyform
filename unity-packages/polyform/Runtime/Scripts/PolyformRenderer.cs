using System;
using System.Collections;
using EliCDavis.Polyform.Artifacts;
using EliCDavis.Polyform.Models;
using UnityEngine;
using UnityEngine.Serialization;

namespace EliCDavis.Polyform
{
    public class PolyformRenderer : MonoBehaviour
    {
        [SerializeField] private Graph graph;

        [SerializeField] private AvailableManifestObject endpoint;

        [SerializeField] private ArtifactLoader[] handlers;

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
            yield return LoadManifest(endpoint.AvailableManifest());
        }

        private IEnumerator LoadManifest(AvailableManifest manifest)
        {
            var manifestsReq = graph.CreateManifest(manifest.Name, manifest.Port);
            yield return manifestsReq.Run();

            foreach (var handler in handlers)
            {
                if (!handler.CanHandle(manifestsReq.Result.Manifest)) continue;
                handler.Handle(graph, manifestsReq.Result);
                yield break;
            }

            throw new Exception($"No handler registered to handle manifest: {manifest.Name}/{manifest.Port}");
        }
    }
}