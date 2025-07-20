using UnityEngine;

namespace EliCDavis.Polyform.Variants
{
    public abstract class Variant<T>: VariantBase
    {
        public override object SampleValue()
        {
            return Sample();
        }

        public abstract T Sample();
    }
}