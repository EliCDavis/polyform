using EliCDavis.Polyform.Requests;
using UnityEngine;

namespace EliCDavis.Polyform
{
    [CreateAssetMenu(fileName = "Connection Config",  menuName = "Polyform/Connection Config", order = 1)]
    public class ConnectionConfig : ScriptableObject
    {
        [SerializeField]
        private string url;

        public GetManifestsRequest AvailableManifests()
        {
            return new GetManifestsRequest(url);
        }
        
        public CreateManifestRequest CreateManifest(string node, string port)
        {
            return new CreateManifestRequest(url, node, port);
        }
        
        public GetProfileRequest Profile()
        {
            return new GetProfileRequest(url);
        }
    }
}
