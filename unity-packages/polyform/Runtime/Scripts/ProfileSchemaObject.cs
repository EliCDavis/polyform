using System;
using System.Collections.Generic;
using EliCDavis.Polyform.Models;
using UnityEngine;

namespace EliCDavis.Polyform
{
    public class ProfileSchemaObject : ScriptableObject
    {
        [Serializable]
        class NamedProperty
        {
            public string Name => name;

            public Property Property => property;

            public NamedProperty(string name, Property property)
            {
                this.name = name;
                this.property = property;
            }

            [SerializeField] private string name;

            [SerializeField] private Property property;
        }

        [SerializeField] private NamedProperty[] properties;

        public void SetData(Dictionary<string, Property> profileResult)
        {
            properties = new NamedProperty[profileResult.Count];

            var i = 0;
            foreach (var keyval in profileResult)
            {
                properties[i] = new NamedProperty(keyval.Key, keyval.Value);
                i++;
            }
        }

        public Dictionary<string, Property> Data()
        {
            var result = new Dictionary<string, Property>();
            if (properties != null)
            {
                foreach (var prop in properties)
                {
                    result[prop.Name] = prop.Property;
                }
            }

            return result;
        }
    }
}