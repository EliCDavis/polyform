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

        private static VisualElement SetupObjectField<T>(VariantProfile profileObject, ObjectField field, string propName)
        {
            var val = profileObject.Get<T>(propName);
            field.objectType = typeof(Variant<T>);
            field.value = val;
            field.RegisterValueChangedCallback((evt) =>
            {
                profileObject.Set(propName, evt.newValue as Variant<T>);
                EditorUtility.SetDirty(profileObject);
            });
            
            // var editor = UnityEditor.Editor.CreateEditor(val);
            // if (editor != null)
            // {
            //     editors.Add(editor);
            //     Debug.Log("Created editor for prop " + propName);
            //     return editor.CreateInspectorGUI();
            // }

            if (val == null)
            {
                return null;
            }
            
            return new InspectorElement(val);

            return null;
        }

        static VisualElement GetProperty(VariantProfile profileObject, string propName, Property prop)
        {
            var root = new VisualElement();
            var objectField = new ObjectField()
            {
                label = propName,
                tooltip = $"{prop}: {prop.Description}",
            };
            root.Add(objectField);

            VisualElement editorEle = null;

            switch (prop.Type)
            {
                case "number":
                    switch (prop.Format)
                    {
                        case "double":
                            editorEle =SetupObjectField<double>(profileObject, objectField, propName);
                            break;

                        default:
                            editorEle =SetupObjectField<float>(profileObject, objectField, propName);
                            break;
                    }

                    break;

                case "integer":
                    editorEle =SetupObjectField<int>(profileObject, objectField, propName);
                    break;

                case "string":
                    switch (prop.Format)
                    {
                        case "color":
                            editorEle =SetupObjectField<Color>(profileObject, objectField, propName);
                            break;

                        default:
                            editorEle =SetupObjectField<string>(profileObject, objectField, propName);
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