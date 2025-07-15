using System.Collections.Generic;
using Newtonsoft.Json;

namespace EliCDavis.Polyform.Models
{
    [System.Serializable]
    public class ManifestEntry
    {
        [JsonProperty("metadata")]
        public Dictionary<string, object> Metadata;
    }
}
