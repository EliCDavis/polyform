using UnityEngine;

namespace EliCDavis.Polyform.Variants.SpawnAreas
{
    public abstract class VolumeSpawnArea : SpawnArea
    {
        [SerializeField] private GroundSettings groundSettings;

        [SerializeField] private Vector3 minScale = Vector3.one;

        [SerializeField] private Vector3 maxScale = Vector3.one;

        [SerializeField] private Vector3 minRotation = Vector3.zero;

        [SerializeField] private Vector3 maxRotation = Vector3.zero;

        protected Color GizmoColor = new Color(1, 0, 0, 0.2f);

        private Vector3 Random(Vector3 min, Vector3 max)
        {
            var range = max - min;
            return min + new Vector3(
                UnityEngine.Random.value * range.x,
                UnityEngine.Random.value * range.y,
                UnityEngine.Random.value * range.z
            );
        }

        public override SpawnTransform SpawnPoint()
        {
            var p = GenerateSpawn();
            var r = Quaternion.Euler(Random(minRotation, maxRotation));
            var s = Random(minScale, maxScale);
            var spawn = new SpawnTransform(p, r, s);

            if (groundSettings == null)
            {
                return spawn;
            }


            return groundSettings.ModifySpawn(this, spawn);
        }

        public abstract Vector3 GenerateSpawn();

        public abstract bool InsideArea(Vector3 p);
    }
}