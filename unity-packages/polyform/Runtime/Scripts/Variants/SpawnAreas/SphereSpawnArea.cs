using UnityEngine;

namespace EliCDavis.Polyform.Variants.SpawnAreas
{
    [AddComponentMenu("Polyform/Variant/Spawn Area/Sphere Spawn Area")]
    public class SphereSpawnArea : SpawnArea
    {
        [SerializeField] private float radius;

        protected override Vector3 GenerateSpawn()
        {
            return transform.position + (Random.insideUnitSphere * radius);
        }

        public override bool InsideArea(Vector3 p)
        {
            return Vector3.Magnitude(transform.position - p) <= radius;
        }
        
        private void OnDrawGizmosSelected()
        {
            Gizmos.color = GizmoColor;
            Gizmos.DrawSphere(transform.position, radius);
        }
    }
}