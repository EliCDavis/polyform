using UnityEngine;

namespace EliCDavis.Polyform.Variants
{
    public abstract class VariantBase: ScriptableObject
    {
        public abstract object SampleValue();
    }
}