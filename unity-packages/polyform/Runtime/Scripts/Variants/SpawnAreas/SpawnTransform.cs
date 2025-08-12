using UnityEngine;

namespace EliCDavis.Polyform.Variants.SpawnAreas
{
    [System.Serializable]
    public class SpawnTransform
    {
        
        public SpawnTransform(Vector3 position)
        {
            this.position = position;
            
            rotation = Quaternion.identity;
            scale = Vector3.one;
        }
        
        public SpawnTransform(Vector3 position, Quaternion rotation, Vector3 scale)
        {
            this.position = position;
            this.rotation = rotation;
            this.scale = scale;
        }

        public void Set(Transform transform)
        {
            transform.position = position;
            transform.rotation = rotation;
            transform.localScale = scale;
        }

        public Vector3 Position => position;

        public Quaternion Rotation => rotation;

        public Vector3 Scale => scale;

        [SerializeField]
        private Vector3 position;
        
        [SerializeField]
        private Quaternion rotation;
        
        [SerializeField]
        private Vector3 scale;
        
    }
}