using System;
using EliCDavis.Polyform.Elements;
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

            root.Add(schemaField);

            var variableContainer = new VisualElement();
            BuildVariables(profile, profile.Schema, variableContainer);
            root.Add(variableContainer);

            schemaField.RegisterValueChangeCallback((evt =>
            {
                BuildVariables(profile, profile.Schema, variableContainer);
            }));

            root.Add(new Button(() =>
            {
                profile.Clear();
                EditorUtility.SetDirty(profile);
            })
            {
                text = "Reset"
            });

            return root;
        }

        private EventCallback<ChangeEvent<T>> SaveValue<T>(ProfileObject profileObject, string prop)
        {
            return evt =>
            {
                profileObject.Set(prop, evt.newValue);
                EditorUtility.SetDirty(profileObject);
            };
        }

        VisualElement CreateField(ProfileObject profileObject, string propName, Property prop)
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
                            dbl.RegisterCallback(SaveValue<double>(profileObject, propName));
                            return dbl;

                        default:
                            var f = new FloatField(propName)
                            {
                                value = profileObject.Get<float>(propName)
                            };
                            f.RegisterCallback(SaveValue<float>(profileObject, propName));
                            return f;
                    }

                case "integer":
                    var i = new IntegerField(propName)
                    {
                        value = (int)profileObject.Get<long>(propName)
                    };
                    // Debug.Log($"Created: {propName}: {profileObject.Get<long>(propName)} ({typeof(long)})");
                    i.RegisterCallback((ChangeEvent<int> evt) =>
                    {
                        profileObject.Set(propName, (long)evt.newValue);
                        EditorUtility.SetDirty(profileObject);
                    });
                    return i;

                case "string":
                    switch (prop.Format)
                    {
                        case "color":
                            ColorUtility.TryParseHtmlString(profileObject.Get<string>(propName), out var color);
                            var colorField = new ColorField(propName)
                            {
                                value = color
                            };
                            colorField.RegisterValueChangedCallback(evt =>
                            {
                                profileObject.Set(propName, "#" + ColorUtility.ToHtmlStringRGBA(evt.newValue));
                                EditorUtility.SetDirty(profileObject);
                            });
                            return colorField;

                        default:
                            var textField = new TextField(propName)
                            {
                                value = profileObject.Get<string>(propName)
                            };
                            textField.RegisterCallback(SaveValue<string>(profileObject, propName));
                            return textField;
                    }

                case "array":
                    switch (prop.Items.Ref)
                    {
                        case "#/definitions/Vector3":
                            var arrayField = new Vector3ArrayField(propName, profileObject.Get<Vector3[]>(propName));
                            arrayField.OnValueChanged += list =>
                            {
                                profileObject.Set(propName, list.ToArray());
                                EditorUtility.SetDirty(profileObject);
                            };
                            return arrayField;                  
                    }
                    break;
          
            }

            throw new Exception($"unimplemented type: {prop.Type}");
        }

        void BuildVariables(ProfileObject profile, ProfileSchemaObject profileSchema, VisualElement root)
        {
            root.Clear();
            if (profile == null || profileSchema == null)
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