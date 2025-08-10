using System;
using System.Collections;
using System.Collections.Generic;
using EliCDavis.Polyform.Editor.Loading;
using Unity.EditorCoroutines.Editor;
using UnityEngine;

namespace EliCDavis.Polyform.Editor
{
    [CreateAssetMenu(fileName = "Manifest Importer", menuName = "Polyform/Manifest Importer", order = 1)]
    public class ManifestImporter : ScriptableObject
    {
        [SerializeField] private AvailableManifestObject endpoint;

        [SerializeField] private ProfileObject profile;

        [SerializeField] private EditorManifestHandler[] handlers;

        public void Import()
        {
            EditorCoroutineUtility.StartCoroutine(LoadManifest(), this);
        }

        private IEnumerator LoadManifest()
        {
            Dictionary<string, object> variableData = null;
            if (profile != null)
            {
                variableData = profile.Profile();
            }

            var manifestsReq = endpoint.Create(variableData);
            yield return manifestsReq.Run();

            foreach (var handler in handlers)
            {
                if (!handler.CanHandle(manifestsReq.Result.Manifest)) continue;
                handler.Handle(endpoint.Graph, manifestsReq.Result, this);
                yield break;
            }

            throw new Exception($"No handler registered to handle manifest: {endpoint.Name}/{endpoint.Port}");
        }
    }
}