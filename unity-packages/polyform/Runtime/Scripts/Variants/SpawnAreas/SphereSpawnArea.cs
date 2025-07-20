using UnityEngine;

namespace EliCDavis.Polyform.Variants.SpawnAreas
{
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
            Gizmos.color = gizmoColor;
            Gizmos.DrawSphere(transform.position, radius);
        }
    }
}