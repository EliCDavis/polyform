using System;
using System.Collections;
using System.Collections.Generic;
using EliCDavis.Polyform.Artifacts;
using EliCDavis.Polyform.Models;
using UnityEngine;

namespace EliCDavis.Polyform
{
    public class PolyformRenderer : MonoBehaviour
    {
        [SerializeField] private Graph graph;

        [SerializeField] private AvailableManifestObject endpoint;

        [SerializeField] private RuntimeArtifactLoader[] handlers;

        [SerializeField] private ProfileObject profile;

        #region Runtime

        private Coroutine running;

        private IRuntimeArtifact runtimeArtifact;

        #endregion


        private void Start()
        {
            Render();
        }

        private void OnEnable()
        {
            if (profile != null)
            {
                profile.OnDataChange += ProfileChange;
            }
        }

        private void OnDisable()
        {
            if (profile != null)
            {
                profile.OnDataChange -= ProfileChange;
            }
        }

        private void Render()
        {
            if (running != null)
            {
                StopCoroutine(running);
            }

            running = StartCoroutine(Run());
        }

        void ProfileChange(string key, object val)
        {
            Render();
        }

        private IEnumerator Run()
        {
            yield return LoadManifest(endpoint.AvailableManifest());
        }

        private IEnumerator LoadManifest(AvailableManifest manifest)
        {
            if (runtimeArtifact != null)
            {
                runtimeArtifact.Unload();
                runtimeArtifact = null;
            }

            Dictionary<string, object> variableData = null;
            if (profile != null)
            {
                variableData = profile.Profile();
            }

            var manifestsReq = graph.CreateManifest(manifest.Name, manifest.Port, variableData);
            yield return manifestsReq.Run();

            foreach (var handler in handlers)
            {
                if (!handler.CanHandle(manifestsReq.Result.Manifest)) continue;
                runtimeArtifact = handler.Handle(gameObject, graph, manifestsReq.Result);
                yield break;
            }

            throw new Exception($"No handler registered to handle manifest: {manifest.Name}/{manifest.Port}");
        }
    }
}