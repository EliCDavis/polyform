using System;
using EliCDavis.Polyform.Models;
using UnityEditor;
using UnityEditor.UIElements;
using UnityEngine;
using UnityEngine.UIElements;

namespace EliCDavis.Polyform.Editor
{
    [CustomEditor(typeof(ProfileObject))]
    public class ProfileObjectEditor : UnityEditor.Editor
    {
        public override VisualElement CreateInspectorGUI()
        {
            var root = new VisualElement();

            var profile = target as ProfileObject;
            if (profile == null)
            {
                return root;
            }

            var schemaField = new PropertyField(serializedObject.FindProperty("schema"));

            // InspectorElement.FillDefaultInspector(root, serializedObject, this);
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

        private EventCallback<ChangeEvent<T>> SaveValue<T>(ProfileObject profileObject, string prop)
        {
            return evt =>
            {
                profileObject.Set(prop, evt.newValue);
                AssetDatabase.SaveAssets();
            };
        }

        BindableElement CreateField(ProfileObject profileObject, string propName, Property prop)
        {
            switch (prop.Type)
            {
                case "number":
                    switch (prop.Format)
                    {
                        case "double":
                            var dbl = new DoubleField(propName)
                            {
                                value = profileObject.Get<double>(propName)
                            };
                            dbl.RegisterValueChangedCallback(SaveValue<double>(profileObject, propName));
                            return dbl;

                        default:
                            var f = new FloatField(propName)
                            {
                                value = profileObject.Get<float>(propName)
                            };
                            f.RegisterValueChangedCallback(SaveValue<float>(profileObject, propName));
                            return f;
                    }

                case "integer":
                    var i = new IntegerField(propName)
                    {
                        value = profileObject.Get<int>(propName)
                    };
                    i.RegisterValueChangedCallback(SaveValue<int>(profileObject, propName));
                    return i;

                case "string":
                    switch (prop.Format)
                    {
                        case "color":
                            ColorUtility.TryParseHtmlString("#" + profileObject.Get<string>(propName), out var color);
                            var colorField = new ColorField(propName)
                            {
                                value = color
                            };
                            colorField.RegisterValueChangedCallback(evt =>
                            {
                                profileObject.Set(propName, ColorUtility.ToHtmlStringRGBA(evt.newValue));
                                AssetDatabase.SaveAssets();
                            });
                            return colorField;

                        default:
                            var textField = new TextField(propName)
                            {
                                value = profileObject.Get<string>(propName)
                            };
                            textField.RegisterValueChangedCallback(SaveValue<string>(profileObject, propName));
                            return textField;
                    }
            }

            throw new Exception($"unimplemented type: {prop.Type}");
        }

        void BuildVariables(ProfileObject profile, ProfileSchemaObject profileSchema, VisualElement root)
        {
            root.Clear();
            if (profile == null)
            {
                return;
            }

            var schema = profileSchema.Data();
            foreach (var keyval in schema)
            {
                var prop = CreateField(profile, keyval.Key, keyval.Value);
                prop.tooltip = keyval.Value.Description;
                root.Add(prop);
            }
        }
    }
}