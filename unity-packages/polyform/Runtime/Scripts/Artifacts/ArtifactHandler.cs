using UnityEngine;

namespace EliCDavis.Polyform.Artifacts
{
    public abstract class ArtifactHandler: ScriptableObject
    {
        public abstract IArtifact Handle(byte[] payload);
    }
}