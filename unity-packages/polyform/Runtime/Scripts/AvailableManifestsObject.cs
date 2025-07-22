using System.Collections.Generic;
using EliCDavis.Polyform.Models;
using EliCDavis.Polyform.Requests;
using UnityEngine;

namespace EliCDavis.Polyform
{
    public class AvailableManifestObject : ScriptableObject
    {
        [SerializeField] private Graph graph;

        [SerializeField] private AvailableManifest availableManifest;

        public string Port => availableManifest.Port;
        
        public string Name => availableManifest.Name;

        public Graph Graph => graph;

        public void SetAvailableManifest(Graph graph, AvailableManifest availableManifest)
        {
            this.graph = graph;
            this.availableManifest = availableManifest;
        }

        public AvailableManifest AvailableManifest()
        {
            return availableManifest;
        }

        public CreateManifestRequest Create(Dictionary<string, object> variableData)
        {
            return graph.CreateManifest(availableManifest.Name, availableManifest.Port, variableData);
        }
    }
}