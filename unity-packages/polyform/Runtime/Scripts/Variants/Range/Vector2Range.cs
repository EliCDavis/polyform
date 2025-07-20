using UnityEngine;

namespace EliCDavis.Polyform.Variants.Range
{
    [CreateAssetMenu(fileName = "Vector2 Range", menuName = "Polyform/Variants/Range/Vector2", order = 1)]
    public class Vector2Range : Variant<Vector2>
    {
        [SerializeField] private Vector2 center;

        [SerializeField] private Vector2 size;

        public override Vector2 Sample()
        {
            return center + new Vector2(
                Random.Range(-size.x, size.x) / 2f,
                Random.Range(-size.y, size.y) / 2f
            );
        }
    }
}