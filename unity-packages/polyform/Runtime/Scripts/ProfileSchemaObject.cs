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
            [SerializeField] public string name;

            [SerializeField] public Property property;
        }

        [SerializeField] private NamedProperty[] properties;

        public void SetData(Dictionary<string,Property> profileResult)
        {
            properties = new NamedProperty[profileResult.Count];

            var i = 0;
            foreach (var keyval in profileResult)
            {
                properties[i] = new NamedProperty()
                {
                    name = keyval.Key,
                    property = keyval.Value
                };
                i++;
            }
        }

        public Dictionary<string, Property> Data()
        {
            var result = new Dictionary<string, Property>();
            foreach (var prop in properties)
            {
                result[prop.name] = prop.property;
            }
            return result;
        }
    }
}