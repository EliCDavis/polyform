using System.Collections.Generic;
using Newtonsoft.Json;

namespace EliCDavis.Polyform.Models
{
    [System.Serializable]
    public class Manifest
    {
        [JsonProperty("main")]
        public string Main { get; set; }

        [JsonProperty("entries")]
        public Dictionary<string, ManifestEntry> Entries { get; set; }
    }
}
