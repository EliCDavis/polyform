using System.Collections.Generic;

namespace EliCDavis.Polyform.Requests
{
    // POST manifest/node/out {profile}
    // GET  manifest/id/file
    // GET  manifests
    // GET  profile

    public class GetProfileRequest : GetJsonRequest<Dictionary<string, string>>
    {
        public GetProfileRequest(string baseUrl) : base(baseUrl)
        {
        }

        protected override string Path => "profile";
    }
}