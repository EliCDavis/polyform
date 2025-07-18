using UnityEngine;

namespace EliCDavis.Polyform.Variants
{
    public abstract class Variant<T>: ScriptableObject
    {
        public abstract T Sample();
    }
}