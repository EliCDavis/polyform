using System.Collections.Generic;
using System.Linq;
using UnityEditor;
using UnityEngine;

namespace EliCDavis.Polyform.Editor
{
    internal static class ScriptableObjectUtil
    {
        public static T AddElement<T>(this ScriptableObject scriptableObject, string name = "Element",
            HideFlags hideFlags = HideFlags.None) where T : ScriptableObject
        {
            T element = ScriptableObject.CreateInstance<T>();

            element.name = name;
            element.hideFlags = hideFlags;

            string scriptableObjectPath = AssetDatabase.GetAssetPath(scriptableObject);

            AssetDatabase.AddObjectToAsset(element, scriptableObjectPath);
            AssetDatabase.SaveAssets();

            Undo.RegisterCreatedObjectUndo(element, "Add element to ScriptableObject");
            return element;
        }

        public static T GetElement<T>(this ScriptableObject scriptableObject) where T : ScriptableObject
        {
            string scriptableObjectPath = AssetDatabase.GetAssetPath(scriptableObject);
            return AssetDatabase.LoadAssetAtPath(scriptableObjectPath, typeof(T)) as T;
        }

        public static List<T> GetElements<T>(this ScriptableObject scriptableObject) where T : ScriptableObject
        {
            string scriptableObjectPath = AssetDatabase.GetAssetPath(scriptableObject);
            return AssetDatabase.LoadAllAssetsAtPath(scriptableObjectPath).OfType<T>().ToList();
        }
    }
}