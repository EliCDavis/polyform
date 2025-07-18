using EliCDavis.Polyform.Models;
using UnityEngine;

namespace EliCDavis.Polyform.Artifacts
{
    public abstract class RuntimeArtifactLoader: ScriptableObject
    {
        public abstract bool CanHandle(Manifest manifest);
        public abstract IRuntimeArtifact Handle(GameObject parent, Graph graph, ManifestInstance manifestInstance);
    }
}