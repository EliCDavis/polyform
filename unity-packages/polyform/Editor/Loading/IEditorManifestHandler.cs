using EliCDavis.Polyform.Models;
using UnityEngine;

namespace EliCDavis.Polyform.Editor.Loading
{
    public abstract class EditorManifestHandler: ScriptableObject
    {
        public abstract bool CanHandle(Manifest manifest);
        public abstract void Handle(Graph graph, ManifestInstance manifestInstance, ScriptableObject scriptableObject);
    }
}