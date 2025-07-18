using UnityEngine;

namespace EliCDavis.Polyform.Variants
{
    [CreateAssetMenu(fileName = "Variants Profile", menuName = "Polyform/Variants/Profile", order = 1)]
    public class VariantProfile: ScriptableObject
    {
        [SerializeField] private ProfileSchemaObject schema;

        [SerializeField] private ScriptableObject[] data;
    }
}