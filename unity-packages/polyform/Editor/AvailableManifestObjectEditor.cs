using UnityEditor;
using UnityEngine.UIElements;

namespace EliCDavis.Polyform.Editor
{
    [CustomEditor(typeof(AvailableManifestObject))]
    public class AvailableManifestObjectEditor : UnityEditor.Editor
    {
        public override VisualElement CreateInspectorGUI()
        {
            var root = new VisualElement();

            var availableManifest = target as AvailableManifestObject;
            if (availableManifest == null)
            {
                return root;
            }

            var available = availableManifest.AvailableManifest();
            var container = new VisualElement
            {
                style =
                {
                    flexDirection = new StyleEnum<FlexDirection>(FlexDirection.Row)
                }
            };
            container.Add(new Label(available.Name));
            container.Add(new Label(available.Port));
            root.Add(container);

            return root;
        }
    }
}