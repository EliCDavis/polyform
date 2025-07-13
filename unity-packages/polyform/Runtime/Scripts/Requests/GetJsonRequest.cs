using System.Text;
using Newtonsoft.Json;
using UnityEngine.Networking;

namespace EliCDavis.Polyform.Requests
{
    public abstract class GetJsonRequest<T>: Request
    {
        public T Result { get; private set; }
        
        protected GetJsonRequest(string baseUrl) : base(baseUrl) { }
        
        protected override string Method => UnityWebRequest.kHttpVerbGET;

        protected override void HandleBody(byte[] data)
        {
            Result = JsonConvert.DeserializeObject<T>(Encoding.UTF8.GetString(data));
        }
    }
}