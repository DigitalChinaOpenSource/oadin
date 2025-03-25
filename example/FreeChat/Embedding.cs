using System;
using System.Collections.Generic;
using System.Net.Http;
using System.Text;
using System.Threading.Tasks;
using Newtonsoft.Json;
using Newtonsoft.Json.Linq;

namespace Core
{
    public class Embedding
    {
        private const string EmbedApiUrl = "http://10.3.74.124:11434/api/embed";
        private const string MilvusApiUrl = "http://your-milvus-url/api/vectors";

        // Process text
        public async Task<List<Segment>> ProcessTextAsync(string text)
        {
            var segments = SplitText(text);

            var vectors = await GetVectorsAsync(segments);
            if (vectors != null)
            {
                var segmentList = new List<Segment>();
                for (int i = 0; i < segments.Count; i++)
                {
                    segmentList.Add(new Segment
                    {
                        Text = segments[i],
                        Vector = vectors[i]
                    });
                }
                return segmentList;
            }
            return null;
        }

        // Split text into segments
        public List<string> SplitText(string text)
        {

            int segmentSize = 100;
            var segments = new List<string>();

            for (int i = 0; i < text.Length; i += segmentSize)
            {
                segments.Add(text.Substring(i, Math.Min(segmentSize, text.Length - i)));
            }

            return segments;
        }

        // Get vectors from embedding service
        public async Task<List<float[]>> GetVectorsAsync(List<string> segments)
        {
            var client = new HttpClient();
            var requestPayload = new
            {
                model = "quentinz/bge-embedding-768",
                input = segments
            };
            var jsonPayload = JsonConvert.SerializeObject(requestPayload);
            var content = new StringContent(jsonPayload, Encoding.UTF8, "application/json");

            var response = await client.PostAsync(EmbedApiUrl, content);
            var responseString = await response.Content.ReadAsStringAsync();

            var responseJson = JObject.Parse(responseString);
            var vectors = responseJson["embeddings"]?.ToObject<List<float[]>>();

            return vectors;
        }

        // Store vectors into Milvus
        public async Task StoreVectorAsync(float[] vector)
        {
            var client = new HttpClient();
            var requestPayload = new { vector = vector };
            var jsonPayload = JsonConvert.SerializeObject(requestPayload);
            var content = new StringContent(jsonPayload, Encoding.UTF8, "application/json");

            await client.PostAsync(MilvusApiUrl, content);
        }
    }

    // Segment class
    public class Segment
    {
        public string Text { get; set; }
        public float[] Vector { get; set; }
    }
}