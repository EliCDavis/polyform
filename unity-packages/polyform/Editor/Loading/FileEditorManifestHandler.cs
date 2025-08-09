using System.Collections;
using System.IO;
using EliCDavis.Polyform.Models;
using Unity.EditorCoroutines.Editor;
using UnityEditor;
using UnityEngine;
using UnityEngine.Networking;

namespace EliCDavis.Polyform.Editor.Loading
{
    [CreateAssetMenu(fileName = "File Download Manifest Handler",
        menuName = "Polyform/Manifest Handlers/Editor/File Download",
        order = 1)]
    public class FileEditorManifestHandler : EditorManifestHandler
    {
        [Tooltip("Folder to save all content to listed in the manifest")] [SerializeField]
        private string folderName;

        public override bool CanHandle(Manifest manifest)
        {
            return true;
        }

        public override void Handle(Graph graph, ManifestInstance manifestInstance,
            ScriptableObject scriptableObject)
        {
            EditorCoroutineUtility.StartCoroutineOwnerless(
                LoadThingCoroutine(graph, manifestInstance, scriptableObject));
        }

        private string ComputeFolderName(ScriptableObject scriptableObject)
        {
            var subFolder = string.IsNullOrWhiteSpace(folderName) ? scriptableObject.name : folderName;
            var currentName = Path.Combine("Assets", subFolder);

            var parentFolder = Path.GetDirectoryName(currentName);
            var newFolderName = Path.GetFileName(currentName);

            Directory.CreateDirectory(parentFolder);
            AssetDatabase.Refresh();
            var guid = AssetDatabase.CreateFolder(parentFolder, newFolderName);
            return AssetDatabase.GUIDToAssetPath(guid);
        }

        private IEnumerator LoadThingCoroutine(Graph graph, ManifestInstance manifestInstance,
            ScriptableObject scriptableObject)
        {
            var folder = ComputeFolderName(scriptableObject);
            foreach (var entry in manifestInstance.Manifest.Entries)
            {
                var url = graph.FormatURl($"manifest/{manifestInstance.Id}/{entry.Key}");
                var request = new UnityWebRequest(url, UnityWebRequest.kHttpVerbGET);

                var path = Path.Combine(folder, entry.Key);
                Directory.CreateDirectory(Path.GetDirectoryName(path));
                request.downloadHandler = new DownloadHandlerFile(path);

                request.SendWebRequest();
                yield return new WaitUntil(() => request.isDone);
            }

            AssetDatabase.Refresh();

            if (!string.IsNullOrWhiteSpace(manifestInstance.Manifest.Main))
            {
                Ping(Path.Combine(folder, manifestInstance.Manifest.Main));
            }
        }

        static void Ping(string assetPath)
        {
            Object asset = AssetDatabase.LoadAssetAtPath<Object>(assetPath);

            if (asset != null)
            {
                EditorGUIUtility.PingObject(asset);
                Selection.activeObject = asset;
            }
        }
    }
}