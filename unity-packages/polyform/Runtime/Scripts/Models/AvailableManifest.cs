using Newtonsoft.Json;
using Newtonsoft.Json.Converters;

namespace EliCDavis.Polyform
{
    [System.Serializable]
    public class AvailableManifest
    {
        [JsonProperty("name")]
        public string Name { get; set; }
        
        [JsonProperty("port")]
        public string Port { get; set; }
    }
}