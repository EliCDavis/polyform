using System;
using System.Collections.Generic;
using EliCDavis.Polyform.Serialization;
using Newtonsoft.Json;
using Newtonsoft.Json.Linq;
using UnityEngine;

namespace EliCDavis.Polyform
{
    [CreateAssetMenu(fileName = "Profile", menuName = "Polyform/Profile", order = 1)]
    public class ProfileObject : ScriptableObject, ISerializationCallbackReceiver
    {
        [Serializable]
        class SerializeEntry
        {
            [SerializeField] public string name;
            [SerializeField] public string type;
            [SerializeField] public string data;
        }

        [SerializeField] private ProfileSchemaObject schema;

        [SerializeField] private SerializeEntry[] serializedData;

        public delegate void DataChanged(string key, object data);

        public event DataChanged OnDataChange;

        private Dictionary<string, object> data;

        public ProfileSchemaObject Schema => schema;

        public void Clear()
        {
            data = new Dictionary<string, object>();
        }

        public void Set(string key, object val)
        {
            // Debug.Log($"Setting {key} to {val} ({val.GetType()})");
            data ??= new Dictionary<string, object>();
            data[key] = val;
            OnDataChange?.Invoke(key, val);
        }

        public T Get<T>(string key)
        {
            if (data == null || !data.ContainsKey(key))
            {
                return default;
            }

            return data[key] is T ? (T)data[key] : default;
        }

        public Dictionary<string, object> Profile()
        {
            return data;
        }

        public void OnBeforeSerialize()
        {

            data ??= new Dictionary<string, object>();

            serializedData = new SerializeEntry[data.Count];
            int i = 0;
            foreach (var keyval in data)
            {
                serializedData[i] = new SerializeEntry()
                {
                    name = keyval.Key,
                    data = JsonConvert.SerializeObject(
                        keyval.Value,
                        Formatting.None,
                        JsonConverters.Converters
                    ),
                    type = keyval.Value.GetType().ToString()
                };
                i++;
            }

            // serializedData = JsonConvert.SerializeObject(
            //     data,
            //     Formatting.None,
            //     new ColorHexConverter(),
            //     new Vector3Converter(),
            //     new Vector2Converter()
            // );

            // Debug.Log($"Saving: {serializedData}");
        }

        public void OnAfterDeserialize()
        {
            data = new Dictionary<string, object>();

            if (serializedData == null)
            {
                return;
            }

            foreach (var prop in serializedData)
            {
                switch (prop.type)
                {
                    case "System.Single":
                        data[prop.name] = JsonConvert.DeserializeObject<float>(prop.data);
                        break;

                    case "System.Double":
                        data[prop.name] = JsonConvert.DeserializeObject<double>(prop.data);
                        break;

                    case "System.String":
                        data[prop.name] = JsonConvert.DeserializeObject<string>(prop.data);
                        break;

                    case "System.Int64":
                        data[prop.name] = JsonConvert.DeserializeObject<long>(prop.data);
                        break;

                    case "System.Int32":
                        data[prop.name] = JsonConvert.DeserializeObject<int>(prop.data);
                        break;

                    case "UnityEngine.Vector2":
                        data[prop.name] = JsonConvert.DeserializeObject<Vector2>(prop.data, new Vector2Converter());
                        break;

                    case "UnityEngine.Vector3":
                        data[prop.name] = JsonConvert.DeserializeObject<Vector3>(prop.data, new Vector3Converter());
                        break;
                    
                    case "UnityEngine.Bounds":
                        data[prop.name] = JsonConvert.DeserializeObject<Bounds>(prop.data, new AABBConverter());
                        break;

                    case "UnityEngine.Vector3[]":
                        data[prop.name] = JsonConvert.DeserializeObject<Vector3[]>(prop.data, new Vector3Converter());
                        break;
                    
                    case "UnityEngine.Vector2[]":
                        data[prop.name] = JsonConvert.DeserializeObject<Vector2[]>(prop.data, new Vector2Converter());
                        break;

                    default:
                        Debug.LogWarning($"Unknown type {prop.type}");
                        break;
                }
            }

            // data = JsonConvert.DeserializeObject<Dictionary<string, object>>(
            //     serializedData,
            //     new ColorHexConverter(),
            //     new Vector3Converter(),
            //     new Vector2Converter()
            // );
            // Debug.Log($"Deserialized: {data["Resolution"]} ({data["Resolution"].GetType()})");
        }

        private static bool LooksLikeVector3(JObject obj)
        {
            return obj.ContainsKey("x") && obj.ContainsKey("y") && obj.ContainsKey("z");
        }
    }
}