using System;
using System.Net.Http;
using System.Text.Json;
using System.Threading.Tasks;

namespace Byze
{
    public class ByzeClient
    {
        private readonly HttpClient _client;
        private readonly string _baseUrl;

        public ByzeClient(string version = "byze/v0.2")
        {
            _baseUrl = $"http://127.0.0.1:16688/{version}";
            _client = new HttpClient { BaseAddress = new Uri(_baseUrl) };
        }

        
        // 通用请求方法
        private async Task<string> RequestAsync(HttpMethod method, string path, object? data = null, Dictionary<string, string>? headers = null)
        {
            try
            {
                if (path.StartsWith("/"))
                {
                    path = path.TrimStart('/');
                }

                HttpRequestMessage request = new HttpRequestMessage(method, $"{_baseUrl}/{path}");

                if (headers != null)
                {
                    foreach (var header in headers)
                    {
                        request.Headers.Add(header.Key, header.Value);
                    }
                }

                if (data != null)
                {
                    var json = JsonSerializer.Serialize(data);
                    request.Content = new StringContent(json, System.Text.Encoding.UTF8, "application/json");
                }

                var response = await _client.SendAsync(request);
                response.EnsureSuccessStatusCode();

                return await response.Content.ReadAsStringAsync();
            }
            catch (Exception ex)
            {
                throw new Exception($"请求 {method} {path} 失败: {ex.Message}");
            }
        }

        // 获取服务
        public async Task<string> GetServicesAsync()
        {
            return await RequestAsync(HttpMethod.Get, "/service");
        }

        // 创建新服务
        public async Task<string> InstallServiceAsync(object data)
        {
            return await RequestAsync(HttpMethod.Post, "/service", data);
        }

        // 更新服务
        public async Task<string> UpdateServiceAsync(object data)
        {
            return await RequestAsync(HttpMethod.Put, "/service", data);
        }

        // 查看模型
        public async Task<string> GetModelsAsync()
        {
            return await RequestAsync(HttpMethod.Get, "/model");
        }

        // 安装模型
        public async Task<string> InstallModelAsync(object data)
        {
            return await RequestAsync(HttpMethod.Post, "/model", data);
        }

        // 流式安装模型
        public async Task InstallModelStreamAsync(
            object data,
            Action<JsonElement> onData,
            Action<string> onError,
            Action onEnd)
        {
            try
            {
                var json = JsonSerializer.Serialize(data);
                var content = new StringContent(json, System.Text.Encoding.UTF8, "application/json");
                var request = new HttpRequestMessage(HttpMethod.Post, "/model/stream")
                {
                    Content = content
                };

                var response = await _client.SendAsync(request, HttpCompletionOption.ResponseHeadersRead);
                response.EnsureSuccessStatusCode();

                using var stream = await response.Content.ReadAsStreamAsync();
                using var reader = new System.IO.StreamReader(stream);

                while (!reader.EndOfStream)
                {
                    var line = await reader.ReadLineAsync();
                    if (string.IsNullOrWhiteSpace(line))
                        continue;

                    try
                    {
                        string rawData = line.StartsWith("data:") ? line.Substring(5) : line;
                        var responseData = JsonSerializer.Deserialize<JsonElement>(rawData);

                        onData?.Invoke(responseData);

                        var status = responseData.GetProperty("status").GetString();
                        if (status == "success" || status == "error")
                        {
                            onEnd?.Invoke();
                            break;
                        }
                    }
                    catch (Exception ex)
                    {
                        onError?.Invoke($"解析流数据失败: {ex.Message}");
                    }
                }
            }
            catch (Exception ex)
            {
                onError?.Invoke($"流式安装模型失败: {ex.Message}");
            }
        }

        // 取消流式安装模型
        public async Task<string> CancelInstallModelAsync(object data){
            return await RequestAsync(HttpMethod.Post, "/model/stream/cancel", data);
        }

        // 删除模型
        public async Task<string> DeleteModelAsync(object data)
        {
            return await RequestAsync(HttpMethod.Delete, "/model", data);
        }

        // 查看模型提供商
        public async Task<string> GetServiceProvidersAsync()
        {
            return await RequestAsync(HttpMethod.Get, "/service_provider");
        }

        // 新增模型提供商
        public async Task<string> AddServiceProviderAsync(object data)
        {
            return await RequestAsync(HttpMethod.Post, "/service_provider", data);
        }

        // 更新模型提供商
        public async Task<string> UpdateServiceProviderAsync(object data)
        {
            return await RequestAsync(HttpMethod.Put, "/service_provider", data);
        }

        // 删除模型提供商
        public async Task<string> DeleteServiceProviderAsync(object data)
        {
            return await RequestAsync(HttpMethod.Delete, "/service_provider", data);
        }

        // 获取模型列表
        public async Task<string> GetModelAvailiableAsync()
        {
            return await RequestAsync(HttpMethod.Get, "/services/models");
        }

        // 获取推荐模型列表
        public async Task<string> GetModelsRecommendedAsync()
        {
            return await RequestAsync(HttpMethod.Get, "/model/recommend");
        }

        // 获取支持模型列表
        public async Task<string> GetModelsSupportedAsync()
        {
            return await RequestAsync(HttpMethod.Get, "/model/support");
        }

        // 获取问学支持模型列表
        public async Task<string> GetSmartvisionModelsSupportedAsync()
        {
            var headers = new Dictionary<string, string>
            {
                { "Custom-Header", "HeaderValue" } // 替换为实际的请求头键值对
            };

            return await RequestAsync(HttpMethod.Get, "/model/support/smartvision", null, headers);
        }


    }
}
