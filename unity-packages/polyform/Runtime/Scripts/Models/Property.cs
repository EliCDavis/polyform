using System;
using Newtonsoft.Json;
using UnityEngine;

namespace EliCDavis.Polyform.Models
{
    [Serializable]
    public class Property
    {
        [SerializeField] [JsonProperty("type")]
        public string Type;

        [SerializeField] [JsonProperty("format")]
        public string Format;

        [SerializeField] [JsonProperty("description")]
        public string Description;
        
        [SerializeField] [JsonProperty("items")]
        public ItemsObject Items;
        
        [SerializeField] [JsonProperty("$ref")]
        public string Ref;
        
        public override string ToString()
        {
            if (Type == "array")
            {
                return $"{Type} ({Items})";
            }
            
            if (string.IsNullOrWhiteSpace(Format))
            {
                return Type;
            }

            return $"{Type} ({Format})";
        }
    }
}