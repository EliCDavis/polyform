using System.Collections.Generic;
using System.Text;
using EliCDavis.Polyform.Models;
using EliCDavis.Polyform.Serialization;
using Newtonsoft.Json;
using UnityEngine.Networking;

namespace EliCDavis.Polyform.Requests
{
    public class CreateManifestRequest : Request
    {
        private string node;

        private string port;

        private Dictionary<string, object> profile;

        public ManifestInstance Result { get; private set; }

        public CreateManifestRequest(string baseUrl, string node, string port,
            Dictionary<string, object> profile = null) : base(baseUrl)
        {
            this.node = node;
            this.port = port;
            this.profile = profile;
        }

        protected override string Method => UnityWebRequest.kHttpVerbPOST;

        protected override string Path => $"manifest/{node}/{port}";

        protected override void HandleResponseBody(byte[] data)
        {
            Result = JsonConvert.DeserializeObject<ManifestInstance>(Encoding.UTF8.GetString(data));
        }

        protected override byte[] RequestBody()
        {
            if (profile == null)
            {
                return null;
            }

            var json = JsonConvert.SerializeObject(profile, Formatting.None, new ColorHexConverter());
            return Encoding.UTF8.GetBytes(json);
        }
    }
}