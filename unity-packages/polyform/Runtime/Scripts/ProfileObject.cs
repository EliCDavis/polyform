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

        public delegate void DataChanged(string key, object data);

        public event DataChanged OnDataChange;

        private Dictionary<string, object> data;

        public ProfileSchemaObject Schema => schema;

        public void Set(string key, object val)
        {
            Debug.Log($"Setting {key} to {val} ({val.GetType()})");
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
            serializedData = JsonConvert.SerializeObject(data);
            Debug.Log($"Saving: {serializedData}");
        }

        public void OnAfterDeserialize()
        {
            Debug.Log($"Loading: {serializedData}");
            data = JsonConvert.DeserializeObject<Dictionary<string, object>>(serializedData);
        }
    }
}