using UnityEngine;

namespace EliCDavis.Polyform.Variants.Set
{
    public class Set<T> : Variant<T>
    {
        [SerializeField] private Variant<T>[] variants;
        
        public override T Sample()
        {
            return variants[Random.Range(0, variants.Length)].Sample();
        }
    }
}