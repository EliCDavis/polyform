using System.Collections.Generic;
using EliCDavis.Polyform.Requests;
using UnityEngine;

namespace EliCDavis.Polyform
{
    [CreateAssetMenu(fileName = "Graph", menuName = "Polyform/Graph", order = 1)]
    public class Graph : ScriptableObject
    {
        [SerializeField] private string url;

        public GetManifestsRequest AvailableManifests()
        {
            return new GetManifestsRequest(url);
        }

        public CreateManifestRequest CreateManifest(string node, string port, Dictionary<string, object> profile = null)
        {
            return new CreateManifestRequest(url, node, port, profile);
        }

        public GetProfileRequest Profile()
        {
            return new GetProfileRequest(url);
        }

        public string FormatURl(string contents)
        {
            return $"{url}/{contents}";
        }
    }
}