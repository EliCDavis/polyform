using EliCDavis.Polyform.Models;
using UnityEngine;

namespace EliCDavis.Polyform.Artifacts
{
    public abstract class ArtifactLoader: ScriptableObject
    {
        public abstract bool CanHandle(Manifest manifest);
        public abstract IArtifact Handle(Graph graph, ManifestInstance manifestInstance);
    }
}