using UnityEngine;

namespace EliCDavis.Polyform.Variants.Range
{
    [CreateAssetMenu(fileName = "Int Range", menuName = "Polyform/Variants/Range/Int", order = 1)]
    public class IntRange : Variant<int>
    {
        [SerializeField] private int min;

        [SerializeField] private int max;

        public override int Sample()
        {
            return Random.Range(min, max);
        }
    }
}