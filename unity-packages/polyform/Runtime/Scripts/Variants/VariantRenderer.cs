using System;
using System.Collections;
using System.Collections.Generic;
using EliCDavis.Polyform.Loading;
using EliCDavis.Polyform.Models;
using EliCDavis.Polyform.Variants.SpawnAreas;
using UnityEngine;

namespace EliCDavis.Polyform.Variants
{
    [AddComponentMenu("Polyform/Variant/Variant Renderer")]
    public class VariantRenderer : MonoBehaviour
    {
        [SerializeField] private AvailableManifestObject endpoint;

        [SerializeField] private RuntimeManifestHandler[] handlers;

        [SerializeField] private VariantProfile profile;

        [SerializeField] private int variantCount;

        [SerializeField] private SpawnArea spawnArea;

        #region Runtime

        private Coroutine running;

        private GameObject[] runtimeGameobjects;

        private IRuntimeManifest[] runtimeArtifacts;

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

            runtimeArtifacts = new IRuntimeManifest[variantCount];
            runtimeGameobjects = new GameObject[variantCount];
            for (var i = 0; i < variantCount; i++)
            {
                yield return LoadManifest(i, profile.GenerateProfile());
            }
        }

        private IEnumerator LoadManifest(int job, Dictionary<string, object> variableData)
        {
            var manifestsReq = endpoint.Create( variableData);
            yield return manifestsReq.Run();

            foreach (var handler in handlers)
            {
                if (!handler.CanHandle(manifestsReq.Result.Manifest)) continue;
                runtimeGameobjects[job] = new GameObject(job.ToString());
                runtimeGameobjects[job].transform.position = spawnArea.SpawnPoint();
                runtimeArtifacts[job] = handler.Handle(runtimeGameobjects[job], endpoint.Graph, manifestsReq.Result);
                yield break;
            }

            throw new Exception($"No handler registered to handle manifest: {endpoint.Name}/{endpoint.Port}");
        }
    }
}