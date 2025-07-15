using UnityEditor;
using UnityEngine;
using UnityEngine.UIElements;

namespace EliCDavis.Polyform.Editor
{
    [CustomEditor(typeof(ProfileSchemaObject))]
    public class ProfileSchemaObjectEditor : UnityEditor.Editor
    {
        public override VisualElement CreateInspectorGUI()
        {
            var padding = new StyleLength(new Length(4, LengthUnit.Pixel));
            var root = new VisualElement();

            var profile = target as ProfileSchemaObject;
            if (profile == null)
            {
                return root;
            }

            var data = profile.Data();
            var i = 0;
            foreach (var keyval in data)
            {
                var container = new VisualElement
                {
                    style =
                    {
                        paddingBottom = padding,
                        paddingLeft = padding,
                        paddingRight = padding,
                        paddingTop = padding,
                        flexDirection = new StyleEnum<FlexDirection>(FlexDirection.Row)
                    }
                };
                if (i % 2 == 1)
                {
                    container.style.backgroundColor = new StyleColor(new Color(1, 1, 1, 0.1f));
                }

                container.Add(new Label(keyval.Key));

                var spacer = new VisualElement();
                spacer.style.flexGrow = 1;
                container.Add(spacer);


                container.Add(new Label(keyval.Value.ToString()));
                root.Add(container);
                i++;
            }


            return root;
        }
    }
}