using System.Collections.Generic;
using Newtonsoft.Json;

namespace EliCDavis.Polyform.Models
{
    public class ProfileSchema
    {
        [JsonProperty("properties")] public Dictionary<string, Property> Properties { get; set; }
    }
}