using UnityEngine;

namespace EliCDavis.Polyform.Variants.Range
{
    [CreateAssetMenu(fileName = "Double Range", menuName = "Polyform/Variants/Range/Double", order = 1)]
    public class DoubleRange : Variant<double>
    {
        [SerializeField] private double min;

        [SerializeField] private double max;

        public override double Sample()
        {
            return Random.Range((float)min, (float)max);
        }
    }
}