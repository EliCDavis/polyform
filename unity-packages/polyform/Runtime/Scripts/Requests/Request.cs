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

        protected abstract void HandleResponseBody(byte[] data);

        protected virtual byte[] RequestBody()
        {
            return null;
        }

        public IEnumerator Run()
        {
            var url = $"{baseUrl}/{Path}";
            var req = new UnityWebRequest(url, Method);
            req.downloadHandler = new DownloadHandlerBuffer();

            var reqBody = RequestBody();
            if (reqBody is { Length: > 0 })
            {
                req.uploadHandler = new UploadHandlerRaw(reqBody);
            }

            yield return req.SendWebRequest();
            if (req.responseCode != 200)
            {
                throw new Exception($"{Method} {url} Returned Response Code {req.responseCode}");
            }

            HandleResponseBody(req.downloadHandler.data);
        }
    }
}