using EliCDavis.Polyform.Models;

namespace EliCDavis.Polyform.Requests
{
    public class GetManifestsRequest : GetJsonRequest<AvailableManifest[]>
    {
        public GetManifestsRequest(string baseUrl) : base(baseUrl)
        {
        }

        protected override string Path => "manifests";
    }
}