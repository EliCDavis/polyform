using System;
using Newtonsoft.Json;
using UnityEngine;

namespace EliCDavis.Polyform.Models
{
    // https://swagger.io/specification/v2/ 
    // Ctrl+F for "Items Object
    [Serializable]
    public class ItemsObject
    {
        [SerializeField] [JsonProperty("$ref")]
        public string Ref;

        [SerializeField] [JsonProperty("type")]
        public string Type;

        [SerializeField] [JsonProperty("format")]
        public string Format;

        public override string ToString()
        {
            if (!string.IsNullOrWhiteSpace(Ref))
            {
                return Ref;
            }

            return !string.IsNullOrWhiteSpace(Format) ? $"{Type} ({Format})" : Type;
        }
    }
}