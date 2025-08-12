using System;
using UnityEngine;

namespace EliCDavis.Polyform.Variants.SpawnAreas
{
    [Serializable]
    public class GroundSettings
    {
        private const int MaxGroundTrys = 100;

        [SerializeField] private bool clampToGround;

        [SerializeField] private float groundInset;

        [SerializeField] private bool alignToGroundNormal;

        public SpawnTransform ModifySpawn(VolumeSpawnArea spawnArea, SpawnTransform spawn)
        {
            if (!clampToGround)
            {
                return spawn;
            }

            var p = spawn.Position;
            var r = spawn.Rotation;
            for (var i = 0; i < MaxGroundTrys; i++)
            {
                if (Physics.Raycast(p, Vector3.down, out var hit))
                {
                    if (spawnArea.InsideArea(hit.point))
                    {
                        p = hit.point - new Vector3(0, groundInset, 0);

                        if (alignToGroundNormal)
                        {
                            Quaternion targetRotation = Quaternion.FromToRotation(Vector3.up, hit.normal);
                            r = targetRotation * r;
                        }
                        
                        break;
                    }
                }

                p = spawnArea.GenerateSpawn();
            }
            
            return new SpawnTransform(
                p,
                r,
                spawn.Scale
            );
        }
    }
}