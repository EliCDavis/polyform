using System.Collections;
using EliCDavis.Polyform.Models;
using Unity.EditorCoroutines.Editor;
using UnityEditor;
using UnityEditor.UIElements;
using UnityEngine.UIElements;

namespace EliCDavis.Polyform.Editor
{
    [CustomEditor(typeof(Graph))]
    public class GraphEditor : UnityEditor.Editor
    {
        public override VisualElement CreateInspectorGUI()
        {
            var root = new VisualElement();

            var graph = target as Graph;
            if (graph == null)
            {
                return root;
            }

            InspectorElement.FillDefaultInspector(root, serializedObject, this);
            root.Add(new Button(RefreshGraphDefinition)
            {
                text = "Refresh Graph Definition"
            });

            return root;
        }

        void RefreshGraphDefinition()
        {
            EditorCoroutineUtility.StartCoroutine(RefreshGraphDefinitionRoutine(), this);
        }

        IEnumerator RefreshGraphDefinitionRoutine()
        {
            var graph = target as Graph;
            if (graph == null)
            {
                yield break;
            }

            var manifestsRequest = graph.AvailableManifests();
            yield return manifestsRequest.Run();

            AlignManifests(graph, manifestsRequest.Result);

            var profileRequest = graph.Profile();
            yield return profileRequest.Run();

            var profile = graph.GetElement<ProfileSchemaObject>();
            if (profile == null)
            {
                profile = graph.AddElement<ProfileSchemaObject>("Variable Profile Schema");
            }

            profile.SetData(profileRequest.Result);
            

            AssetDatabase.SaveAssets();
        }

        private void AlignManifests(Graph graph, AvailableManifest[] availableManifests)
        {
            var availableManifestObjects = graph.GetElements<AvailableManifestObject>();

            var keep = new bool[availableManifestObjects.Count];
            foreach (var available in availableManifests)
            {
                var name = $"Endpoint: {available.Name} {available.Port}";
                AvailableManifestObject found = null;
                for (var i = 0; i < availableManifestObjects.Count; i++)
                {
                    if (availableManifestObjects[i].name != name) continue;
                    found = availableManifestObjects[i];
                    keep[i] = true;
                }

                if (found == null)
                {
                    found = graph.AddElement<AvailableManifestObject>(name);
                }
                
                found.SetAvailableManifest(available);
            }

            for (var i  = 0; i < availableManifestObjects.Count; i ++)
            {
                if (keep[i])
                {
                    continue;
                }
                AssetDatabase.RemoveObjectFromAsset(availableManifestObjects[i]);
            }
        }
    }
}