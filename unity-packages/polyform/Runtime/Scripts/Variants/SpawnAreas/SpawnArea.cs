using UnityEngine;

namespace EliCDavis.Polyform.Variants.SpawnAreas
{
    public abstract class SpawnArea : MonoBehaviour
    {
        [SerializeField] private bool clampToGround;

        [SerializeField] private float groundInset;

        protected Color gizmoColor = new Color(1, 0, 0, 0.2f);
        
        public Vector3 SpawnPoint()
        {
            var p = GenerateSpawn();
            if (!clampToGround)
            {
                return p;
            }

            for (int i = 0; i < 100; i++)
            {
                if (Physics.Raycast(p, Vector3.down, out var hit))
                {
                    if (InsideArea(hit.point))
                    {
                        return hit.point - new Vector3(0, groundInset, 0);
                    }
                }
                p = GenerateSpawn();
            }

            return p;
        }

        protected abstract Vector3 GenerateSpawn();

        public abstract bool InsideArea(Vector3 p);
    }
}