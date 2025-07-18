using UnityEngine;

namespace EliCDavis.Polyform.Variants.Range
{
    [CreateAssetMenu(fileName = "Color Range", menuName = "Polyform/Variants/Range/Color", order = 1)]
    public class ColorRange : Variant<Color>
    {
        [SerializeField] private Color min;

        [SerializeField] private Color max;

        public override Color Sample()
        {
            return Color.Lerp(min, max, Random.Range(0f, 1f));
        }
    }
}