using UnityEngine;

namespace EliCDavis.Polyform.Variants.SpawnAreas
{
    [AddComponentMenu("Polyform/Variant/Spawn Area/Box Spawn Area")]
    public class BoxSpawnArea : VolumeSpawnArea
    {
        private bool InRange(float v, float min, float max)
        {
            return v >= min && v <= max;
        }

        public override Vector3 GenerateSpawn()
        {
            return transform.TransformPoint(new Vector3(
                Random.Range(-0.5f, 0.5f),
                Random.Range(-0.5f, 0.5f),
                Random.Range(-0.5f, 0.5f)
            ));
        }

        public override bool InsideArea(Vector3 p)
        {
            var inverse = transform.InverseTransformPoint(p);
            return InRange(inverse.x, -0.5f, 0.5f) && 
                   InRange(inverse.y, -0.5f, 0.5f) &&
                   InRange(inverse.z, -0.5f, 0.5f);
        }

        private void OnDrawGizmosSelected()
        {
            var original = Gizmos.matrix;
            Gizmos.matrix = transform.localToWorldMatrix;
            Gizmos.color = GizmoColor;
            Gizmos.DrawCube(Vector3.zero, Vector3.one);
            Gizmos.matrix = original;
        }
    }
}