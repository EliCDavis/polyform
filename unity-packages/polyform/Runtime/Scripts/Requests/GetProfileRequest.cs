using System.Collections.Generic;
using EliCDavis.Polyform.Models;

namespace EliCDavis.Polyform.Requests
{
    // POST manifest/node/out {profile}
    // GET  manifest/id/file
    // GET  manifests
    // GET  profile

    public class GetProfileRequest : GetJsonRequest<Dictionary<string, Property> >
    {
        public GetProfileRequest(string baseUrl) : base(baseUrl)
        {
        }

        protected override string Path => "profile";
    }
}