using Newtonsoft.Json;

namespace EliCDavis.Polyform.Models
{
    [System.Serializable]
    public class ManifestInstance
    {
        [JsonProperty("manifest")] public Manifest Manifest { get; set; }

        [JsonProperty("id")] public string Id { get; set; }
    }
}