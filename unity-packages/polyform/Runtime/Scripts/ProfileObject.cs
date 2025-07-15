using System.Collections.Generic;
using Newtonsoft.Json;
using UnityEngine;

namespace EliCDavis.Polyform
{
    [CreateAssetMenu(fileName = "Profile", menuName = "Polyform/Profile", order = 1)]
    public class ProfileObject : ScriptableObject, ISerializationCallbackReceiver
    {
        [SerializeField] private ProfileSchemaObject schema;

        [SerializeField] private string serializedData;

        private Dictionary<string, object> data;

        public ProfileSchemaObject Schema => schema;

        public void Set(string key, object val)
        {
            data ??= new Dictionary<string, object>();
            data[key] = val;
        }

        public T Get<T>(string key)
        {
            if (data == null || !data.ContainsKey(key))
            {
                return default;
            }

            return data[key] is T ? (T)data[key] : default;
        }

        public void OnBeforeSerialize()
        {
            serializedData = JsonConvert.SerializeObject(data);
        }

        public void OnAfterDeserialize()
        {
            data = JsonConvert.DeserializeObject<Dictionary<string, object>>(serializedData);
        }
    }
}