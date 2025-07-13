using UnityEngine.Networking;

namespace EliCDavis.Polyform.Requests
{
    public class GetManifestEntryRequest : Request
    {
        private readonly string manifestId;

        private readonly string entry;

        public byte[] Result { get; private set; }

        public GetManifestEntryRequest(string baseUrl, string manifestId, string entry) : base(baseUrl)
        {
            this.manifestId = manifestId;
            this.entry = entry;
        }

        protected override string Method => UnityWebRequest.kHttpVerbGET;

        protected override string Path => $"manifest/{manifestId}/{entry}";

        protected override void HandleBody(byte[] data)
        {
            Result = data;
        }
    }
}