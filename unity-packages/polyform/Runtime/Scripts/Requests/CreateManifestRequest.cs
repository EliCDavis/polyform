using System.Text;
using Newtonsoft.Json;
using UnityEngine.Networking;

namespace EliCDavis.Polyform.Requests
{
    public class CreateManifestRequest: Request
    {
        private string node;
        
        private string port;
        
        public CreateManifestResponse Result { get; private set; }
        
        public CreateManifestRequest(string baseUrl, string node, string port) : base(baseUrl)
        {
            this.node = node;
            this.port = port;
        }

        protected override string Method => UnityWebRequest.kHttpVerbPOST;

        protected override string Path => $"manifest/{node}/{port}";
        
        protected override void HandleBody(byte[] data)
        {
            Result = JsonConvert.DeserializeObject<CreateManifestResponse>(Encoding.UTF8.GetString(data));
        }
    }
}