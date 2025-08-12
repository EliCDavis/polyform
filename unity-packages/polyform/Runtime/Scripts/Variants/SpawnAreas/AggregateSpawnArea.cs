using EliCDavis.Polyform.Utils;
using UnityEngine;

namespace EliCDavis.Polyform.Variants.SpawnAreas
{
    [AddComponentMenu("Polyform/Variant/Spawn Area/Aggregate Spawn Area")]
    public class AggregateSpawnArea : SpawnArea
    {
        [SerializeField] private WeightedListConfig<SpawnArea> spawnAreas;

        private WeightedList<SpawnArea> areas;

        private void Awake()
        {
            areas = spawnAreas.List();
        }

        public override SpawnTransform SpawnPoint()
        {
            if (areas.Count == 0)
            {
                throw new System.InvalidOperationException("can't generate spawn without children spawners");
            }

            return areas.Next().SpawnPoint();
        }
    }
}