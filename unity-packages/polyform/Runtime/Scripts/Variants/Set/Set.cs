using EliCDavis.Polyform.Utils;
using UnityEngine;

namespace EliCDavis.Polyform.Variants.Set
{
    public class Set<T> : Variant<T>
    {
        [SerializeField] private WeightedListConfig<Variant<T>> variants;
        
        public override T Sample()
        {
            return variants.List().Next().Sample();
        }
    }
}