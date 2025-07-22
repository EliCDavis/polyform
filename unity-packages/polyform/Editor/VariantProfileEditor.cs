using System;
using System.Collections.Generic;
using EliCDavis.Polyform.Models;
using EliCDavis.Polyform.Variants;
using UnityEditor;
using UnityEditor.UIElements;
using UnityEngine.UIElements;
using UnityEngine;

namespace EliCDavis.Polyform.Editor
{
    [CustomEditor(typeof(VariantProfile))]
    public class VariantProfileEditor : UnityEditor.Editor
    {
        public override VisualElement CreateInspectorGUI()
        {
            var root = new VisualElement();
            var profile = target as VariantProfile;
            if (profile == null)
            {
                return root;
            }

            var schemaField = new PropertyField(serializedObject.FindProperty("schema"));
            root.Add(schemaField);

            var variableContainer = new VisualElement();
            BuildVariables(profile, profile.Schema, variableContainer);
            root.Add(variableContainer);

            schemaField.RegisterValueChangeCallback((evt =>
            {
                BuildVariables(profile, profile.Schema, variableContainer);
            }));

            return root;
        }

        private static VisualElement SetupObjectField<T>(VariantProfile profileObject, ObjectField field,
            string propName)
        {
            var container = new VisualElement();
            var val = profileObject.Get<T>(propName);
            field.objectType = typeof(Variant<T>);
            field.value = val;
            field.RegisterValueChangedCallback((evt) =>
            {
                container.Clear();

                var newVal = evt.newValue as Variant<T>;
                profileObject.Set(propName, newVal);
                EditorUtility.SetDirty(profileObject);

                if (newVal != null)
                {
                    container.Add(new InspectorElement(newVal));
                }
            });

            if (val != null)
            {
                container.Add(new InspectorElement(val));
            }

            return container;
        }

        static VisualElement GetProperty(VariantProfile profileObject, string propName, Property prop)
        {
            var root = new VisualElement();
            var tooltip = string.IsNullOrWhiteSpace(prop.Description) ? prop.ToString() : $"{prop}: {prop.Description}";
            var objectField = new ObjectField()
            {
                label = propName,
                tooltip = tooltip,
            };
            root.Add(objectField);

            VisualElement editorEle = null;

            switch (prop.Type)
            {
                case "number":
                    switch (prop.Format)
                    {
                        case "double":
                            editorEle = SetupObjectField<double>(profileObject, objectField, propName);
                            break;

                        default:
                            editorEle = SetupObjectField<float>(profileObject, objectField, propName);
                            break;
                    }

                    break;

                case "integer":
                    editorEle = SetupObjectField<int>(profileObject, objectField, propName);
                    break;

                case "string":
                    switch (prop.Format)
                    {
                        case "color":
                            editorEle = SetupObjectField<Color>(profileObject, objectField, propName);
                            break;

                        default:
                            editorEle = SetupObjectField<string>(profileObject, objectField, propName);
                            break;
                    }

                    break;
                default:
                    throw new Exception($"unimplemented type: {prop.Type}");
            }

            if (editorEle != null)
            {
                root.Add(editorEle);
            }

            return root;
        }

        void BuildVariables(VariantProfile profile, ProfileSchemaObject profileSchema, VisualElement root)
        {
            root.Clear();

            if (profile == null || profileSchema == null)
            {
                return;
            }

            var schema = profileSchema.Data();
            foreach (var keyval in schema)
            {
                root.Add(GetProperty(profile, keyval.Key, keyval.Value));
            }
        }
    }
}