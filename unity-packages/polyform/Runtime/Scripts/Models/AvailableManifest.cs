using Newtonsoft.Json;

namespace EliCDavis.Polyform.Models
{
    [System.Serializable]
    public class AvailableManifest
    {
        [JsonProperty("name")] public string Name;

        [JsonProperty("port")] public string Port;
    }
}