using System;
using System.Collections.Generic;
using UnityEngine;

namespace EliCDavis.Polyform.Variants
{
    [CreateAssetMenu(fileName = "Variants Profile", menuName = "Polyform/Variants/Profile", order = 1)]
    public class VariantProfile : ScriptableObject
    {
        [Serializable]
        private class NamedReference
        {
            [SerializeField] public string name;
            [SerializeField] public VariantBase reference;
        }

        [SerializeField] private ProfileSchemaObject schema;

        [SerializeField] private NamedReference[] references;

        public ProfileSchemaObject Schema => schema;

        public delegate void DataChanged(string key, object data);

        public event DataChanged OnDataChange;

        public Dictionary<string, object> GenerateProfile()
        {
            InitReferences();
            var result = new Dictionary<string, object>();
            foreach (var reference in references)
            {
                if (reference.reference == null)
                {
                    continue;
                }

                result[reference.name] = reference.reference.SampleValue();
            }

            return result;
        }

        void CreateReferences()
        {
            var data = schema.Data();
            references = new NamedReference[data.Count];
            var i = 0;
            foreach (var keyval in data)
            {
                references[i] = new NamedReference()
                {
                    name = keyval.Key,
                };
                i++;
            }
        }
        
        private void InitReferences()
        {
            if (schema == null)
            {
                references = null;
                return;
            }

            if (references == null)
            {
                Debug.Log("creating...");
                CreateReferences();
                return;
            }

            // Check to make sure we're currently alligned with the schema
            var data = schema.Data();
            if (data.Count != references.Length)
            {
                CreateReferences();
                return;
            }

            foreach (var namedRef in references)
            {
                if (data.ContainsKey(namedRef.name)) continue;
                CreateReferences();
                return;
            }
        }

        public void Set<T>(string key, Variant<T> val)
        {
            // Debug.Log($"Setting {key} to {val} ({val.GetType()})");
            InitReferences();
            foreach (var reference in references)
            {
                if (reference.name != key) continue;
                reference.reference = val;
                OnDataChange?.Invoke(key, val);
                Debug.Log("Set!!!" + key);
                return;
            }

            throw new Exception($"Schema doesn't contain variable {key}");
        }

        public Variant<T> Get<T>(string key)
        {
            InitReferences();
            foreach (var reference in references)
            {
                if (reference.name != key) continue;
                return reference.reference is Variant<T> variant ? variant : default;
            }

            throw new Exception($"Schema doesn't contain variable {key}");
        }
    }
}