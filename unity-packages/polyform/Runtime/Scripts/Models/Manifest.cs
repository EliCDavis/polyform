using System.Collections.Generic;

namespace EliCDavis.Polyform
{
    [System.Serializable]
    public class Manifest
    {
        public string Main { get; set; }

        public Dictionary<string, ManifestEntry> Entries { get; set; }
    }
}
