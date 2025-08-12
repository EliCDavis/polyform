using UnityEngine;

namespace EliCDavis.Polyform.Variants.SpawnAreas
{
    public abstract class SpawnArea : MonoBehaviour
    {
        public abstract SpawnTransform SpawnPoint();
    }
}