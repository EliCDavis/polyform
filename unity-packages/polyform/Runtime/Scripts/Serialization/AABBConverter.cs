using System;
using Newtonsoft.Json;
using UnityEngine;

namespace EliCDavis.Polyform.Serialization
{
    public class AABBConverter: JsonConverter
    {
        public override bool CanConvert(Type objectType)
        {
            return objectType == typeof(Bounds);
        }

        public override object ReadJson(JsonReader reader, Type objectType, object existingValue,
            JsonSerializer serializer)
        {
            var t = serializer.Deserialize(reader);
            var iv = JsonConvert.DeserializeObject<Bounds>(t.ToString());
            return iv;
        }

        public override void WriteJson(JsonWriter writer, object value, JsonSerializer serializer)
        {
            Bounds v = (Bounds)value;

            writer.WriteStartObject();

            var center = v.center;
            writer.WritePropertyName("center");
            writer.WriteStartObject();
            writer.WritePropertyName("x");
            writer.WriteValue(center.x);
            writer.WritePropertyName("y");
            writer.WriteValue(center.y);
            writer.WritePropertyName("z");
            writer.WriteValue(center.z);
            writer.WriteEndObject();
            
            var extents = v.extents;
            writer.WritePropertyName("extents");
            writer.WriteStartObject();
            writer.WritePropertyName("x");
            writer.WriteValue(extents.x);
            writer.WritePropertyName("y");
            writer.WriteValue(extents.y);
            writer.WritePropertyName("z");
            writer.WriteValue(extents.z);
            writer.WriteEndObject();
            
            writer.WriteEndObject();
        }
    }
}