using EliCDavis.Polyform.Utils;
using UnityEngine;

namespace EliCDavis.Polyform.Variants.SpawnAreas
{
    public class AggregateSpawnArea : SpawnArea
    {
        [SerializeField] private WeightedListConfig<SpawnArea> spawnAreas;

        private WeightedList<SpawnArea> areas;

        private void Awake()
        {
            areas = spawnAreas.List();
        }

        protected override Vector3 GenerateSpawn()
        {
            if (areas.Count == 0)
            {
                throw new System.InvalidOperationException("can't generate spawn without children spawners");
            }

            return areas.Next().SpawnPoint();
        }

        public override bool InsideArea(Vector3 p)
        {
            foreach (var area in areas)
            {
                if (area.InsideArea(p))
                {
                    return true;
                }
            }

            return false;
        }
    }
}