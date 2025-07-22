using System.Text;
using Newtonsoft.Json;
using UnityEngine;
using UnityEngine.Networking;

namespace EliCDavis.Polyform.Requests
{
    public abstract class GetJsonRequest<T>: Request
    {
        public T Result { get; private set; }
        
        protected GetJsonRequest(string baseUrl) : base(baseUrl) { }
        
        protected override string Method => UnityWebRequest.kHttpVerbGET;

        protected override void HandleResponseBody(byte[] data)
        {
            JsonSerializerSettings settings = new JsonSerializerSettings();
            
            // Ignore is for handling $ in property field names (%ref from swagger)
            settings.MetadataPropertyHandling = MetadataPropertyHandling.Ignore;            

            Result = JsonConvert.DeserializeObject<T>(Encoding.UTF8.GetString(data), settings);
            // Debug.Log(Encoding.UTF8.GetString(data));
        }
    }
}