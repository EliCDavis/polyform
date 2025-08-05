using EliCDavis.Polyform.Models;
using UnityEngine;

namespace EliCDavis.Polyform.Loading
{
    public abstract class RuntimeManifestHandler: ScriptableObject
    {
        public abstract bool CanHandle(Manifest manifest);
        public abstract IRuntimeManifest Handle(GameObject parent, Graph graph, ManifestInstance manifestInstance);
    }
}