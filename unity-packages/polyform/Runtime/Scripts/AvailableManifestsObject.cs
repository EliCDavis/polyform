using EliCDavis.Polyform.Models;
using UnityEngine;

namespace EliCDavis.Polyform
{
    public class AvailableManifestObject: ScriptableObject
    {
        [SerializeField] private AvailableManifest availableManifest;

        public void SetAvailableManifest(AvailableManifest availableManifest)
        {
            this.availableManifest = availableManifest;
        }

        public AvailableManifest AvailableManifest()
        {
            return availableManifest;
        }

    }
}