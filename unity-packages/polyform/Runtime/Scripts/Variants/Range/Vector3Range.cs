using UnityEngine;

namespace EliCDavis.Polyform.Variants.Range
{
    [CreateAssetMenu(fileName = "Vector3 Range", menuName = "Polyform/Variants/Range/Vector3", order = 1)]
    public class Vector3Range : Variant<Vector3>
    {
        [SerializeField] private Vector3 center;

        [SerializeField] private Vector3 size;

        public override Vector3 Sample()
        {
            return center + new Vector3(
                Random.Range(-size.x, size.x),
                Random.Range(-size.y, size.y),
                Random.Range(-size.z, size.z)
            );
        }
    }
}