using UnityEngine;

namespace EliCDavis.Polyform.Variants.Range
{
    [CreateAssetMenu(fileName = "Float Range", menuName = "Polyform/Variants/Range/Float", order = 1)]
    public class FloatRange : Variant<float>
    {
        [SerializeField] private float min;

        [SerializeField] private float max;

        public override float Sample()
        {
            return Random.Range(min, max);
        }
    }
}