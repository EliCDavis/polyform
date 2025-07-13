using System;
using System.Collections;
using UnityEngine.Networking;

namespace EliCDavis.Polyform.Requests
{
    public abstract class Request
    {
        private readonly string baseUrl;

        protected Request(string baseUrl)
        {
            this.baseUrl = baseUrl;
        }

        protected abstract string Method { get; }
        protected abstract string Path { get; }

        protected abstract void HandleBody(byte[] data);

        public IEnumerator Run()
        {
            var url = $"{baseUrl}/{Path}";
            var req = new UnityWebRequest(url, Method);
            req.downloadHandler = new DownloadHandlerBuffer();
            yield return req.SendWebRequest();
            if (req.responseCode != 200)
            {
                throw new Exception($"{Method} {url} Returned Response Code {req.responseCode}");
            }

            HandleBody(req.downloadHandler.data);
        }
    }
}