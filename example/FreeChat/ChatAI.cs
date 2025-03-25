using System;
using System.Collections.Generic;
using System.Net.Http;
using System.Text;
using System.Threading.Tasks;
using Newtonsoft.Json;
using Newtonsoft.Json.Linq;

namespace Core
{
    public class ChatAI
    {
        public List<object> ChatHistory { get; private set; }

        public const bool Stream = true;

        public string ApiUrl { get; set; } = "http://localhost:16688/byze/v0.2/services/chat";

        public string ApiKey { get; set; } = string.Empty;

        public string ModelName = "deepseek-r1";
        public ChatAI()
        {
            this.ChatHistory = new List<object>();
        }

        // Encode the chat history into request payload
        public async Task<string> Respond(string content)
        {
            if (string.IsNullOrWhiteSpace(content))
                return "请输入内容。";

            ChatHistory.Add(new { role = "user", content });

            var response = await GetAIResponse(content);

            ChatHistory.Add(new { role = "assistant", content = response });

            return response;
        }

        // Get and Parse AI response
        private async Task<string> GetAIResponse(string question)
        {
            var client = new HttpClient();
            var requestPayload = new
            {
                model = ModelName,
                messages = ChatHistory,
                stream = Stream
            };

            var jsonPayload = JsonConvert.SerializeObject(requestPayload);
            var content = new StringContent(jsonPayload, Encoding.UTF8, "application/json");

            if (!string.IsNullOrEmpty(ApiKey))
            {
                client.DefaultRequestHeaders.Add("Authorization", $"Bearer {ApiKey}");
            }

            Console.WriteLine(ApiUrl);
            try
            {
                var response = await client.PostAsync(ApiUrl, content);
                response.EnsureSuccessStatusCode();

                var responseString = await response.Content.ReadAsStringAsync();
                var responseJson = JObject.Parse(responseString);

                string contentField;

                var uri = new Uri(ApiUrl);
                if (uri.Host == "127.0.0.1" || uri.Host == "localhost")
                {
                    contentField = responseJson["message"]?["content"]?.ToString();
                }
                else
                {
                    contentField = responseJson["choices"]?[0]?["message"]?["content"]?.ToString().Trim();
                }

                return contentField ?? "未能获取到AI的回答";
            }
            catch (Exception ex)
            {
                Console.WriteLine($"请求AI服务时发生异常: {ex.Message}");
                return "AI服务异常";
            }
        }
    }
}