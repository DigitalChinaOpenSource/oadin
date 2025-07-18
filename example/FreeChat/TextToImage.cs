using System;
using System.Collections.Generic;
using System.Net.Http;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;
//using Newtonsoft.Json;
//using Newtonsoft.Json.Linq;

namespace Core
{
    public class TextToImageRequest
    {
        public string? Model { get; set; }
        public bool? Stream { get; set; }
        public InputData Input { get; set; } = new InputData();

        public class InputData
        {
            public string Prompt { get; set; } = string.Empty;
        }
    }
    public class TextToImageResponse
    {
        public ResponseData? Data { get; set; }
        public string? Id { get; set; }

        public class ResponseData
        {
            public string? Url { get; set; }
        }
    }

    // Get and Parse AI response
    public class TextToImageClient
    {
        private readonly HttpClient _httpClient;
        public string ModelName { get; set; } = "wanx2.1-t2i-turbo";
        private const string ApiUrl = "https://127.0.0.1:16688/oadin/v0.2/services/text-to-image";
        private const string ApiKey = "your-api-key-here";

        public TextToImageClient()
        {
            _httpClient = new HttpClient();
            _httpClient.DefaultRequestHeaders.Add("Authorization", $"Bearer {ApiKey}");
        }
        public async Task<TextToImageResponse> GenerateImageAsync(string prompt)
        {
            var request = new TextToImageRequest
            {
                Model = ModelName,
                Input = new TextToImageRequest.InputData
                {
                    Prompt = prompt
                }
            };

            var json = JsonSerializer.Serialize(request);
            var content = new StringContent(json, Encoding.UTF8, "application/json");

            var response = await _httpClient.PostAsync(ApiUrl, content);
            response.EnsureSuccessStatusCode();

            var responseJson = await response.Content.ReadAsStringAsync();
            return JsonSerializer.Deserialize<TextToImageResponse>(responseJson)
                ?? throw new Exception("Failed to deserialize response");
        }
    }
}