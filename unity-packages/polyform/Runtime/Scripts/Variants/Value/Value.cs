using UnityEngine;

namespace EliCDavis.Polyform.Variants.Value
{
    public class Value<T> : Variant<T>
    {
        [SerializeField] private T value;

        public override T Sample()
        {
            return value;
        }
    }
}