using Newtonsoft.Json;

namespace EliCDavis.Polyform.Serialization
{
    public static class JsonConverters
    {
        private static JsonConverter[] converters;
        
        public static JsonConverter[] Converters
        {
            get
            {
                return converters ??= new JsonConverter[]
                {
                    new Vector2Converter(),
                    new Vector3Converter(),
                    new ColorHexConverter(),
                    new AABBConverter()
                };
            }
        }
    }
}