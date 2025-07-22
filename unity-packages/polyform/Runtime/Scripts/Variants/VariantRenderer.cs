using System;
using System.Collections;
using System.Collections.Generic;
using EliCDavis.Polyform.Artifacts;
using EliCDavis.Polyform.Models;
using EliCDavis.Polyform.Variants.SpawnAreas;
using UnityEngine;

namespace EliCDavis.Polyform.Variants
{
    public class VariantRenderer : MonoBehaviour
    {
        [SerializeField] private Graph graph;

        [SerializeField] private AvailableManifestObject endpoint;

        [SerializeField] private RuntimeArtifactLoader[] handlers;

        [SerializeField] private VariantProfile profile;

        [SerializeField] private int variantCount;

        [SerializeField] private SpawnArea spawnArea;

        #region Runtime

        private Coroutine running;

        private GameObject[] runtimeGameobjects;

        private IRuntimeArtifact[] runtimeArtifacts;

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
            if (runtimeArtifacts != null)
            {
                foreach (var artifact in runtimeArtifacts)
                {
                    artifact.Unload();
                }

                foreach (var go in runtimeGameobjects)
                {
                    Destroy(go);
                }
            }

            runtimeArtifacts = new IRuntimeArtifact[variantCount];
            runtimeGameobjects = new GameObject[variantCount];
            for (int i = 0; i < variantCount; i++)
            {
                yield return LoadManifest(i, endpoint.AvailableManifest(), profile.GenerateProfile());
            }
        }

        private IEnumerator LoadManifest(int job, AvailableManifest manifest, Dictionary<string, object> variableData)
        {
            var manifestsReq = graph.CreateManifest(manifest.Name, manifest.Port, variableData);
            yield return manifestsReq.Run();

            foreach (var handler in handlers)
            {
                if (!handler.CanHandle(manifestsReq.Result.Manifest)) continue;
                runtimeGameobjects[job] = new GameObject(job.ToString());
                runtimeGameobjects[job].transform.position = spawnArea.SpawnPoint();
                runtimeArtifacts[job] = handler.Handle(runtimeGameobjects[job], graph, manifestsReq.Result);
                yield break;
            }

            throw new Exception($"No handler registered to handle manifest: {manifest.Name}/{manifest.Port}");
        }
    }
}