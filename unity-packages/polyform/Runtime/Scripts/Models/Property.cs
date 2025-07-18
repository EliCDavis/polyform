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
        
        public override string ToString()
        {
            if (string.IsNullOrWhiteSpace(Format))
            {
                return Type;
            }

            return $"{Type} ({Format})";
        }
    }
}