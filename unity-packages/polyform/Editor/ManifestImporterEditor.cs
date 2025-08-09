using UnityEditor;
using UnityEditor.UIElements;
using UnityEngine.UIElements;

namespace EliCDavis.Polyform.Editor
{
    [CustomEditor(typeof(ManifestImporter))]
    public class ManifestImporterEditor : UnityEditor.Editor
    {
        
        public override VisualElement CreateInspectorGUI()
        {
            var root = new VisualElement();
            var manifestImporter = target as ManifestImporter;
            if (manifestImporter == null)
            {
                return root;
            }
            InspectorElement.FillDefaultInspector(root, serializedObject, this);

            root.Add(new Button(manifestImporter.Import)
            {
                text = "Run"
            });

            return root;
        }
        
    }
}