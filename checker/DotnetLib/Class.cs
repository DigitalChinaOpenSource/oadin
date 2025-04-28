using System;
using System.Diagnostics;
using System.Threading;
using System.Threading.Tasks;
using Microsoft.Win32;
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
        public async Task<string> GetSmartvisionModelsSupportedAsync(object data)
        {
            var headers = data as Dictionary<string, string>;

            return await RequestAsync(HttpMethod.Get, "/model/support/smartvision", null, headers);
        }

        // 导入配置文件
        public async Task<string> ImportConfigAsync(string filePath)
        {
            var data = new { file_path = filePath };
            return await RequestAsync(HttpMethod.Post, "/config/import", data);
        }

        // 导出配置文件
        public async Task<string> ExportConfigAsync(object data)
        {
            try
            {
                // 调用 RequestAsync 获取配置文件的 JSON 响应
                var config = await RequestAsync(HttpMethod.Get, "/config/export", data);

                // 获取用户目录
                string userDirectory = Environment.GetFolderPath(Environment.SpecialFolder.UserProfile);

                // 构造 Byze 文件夹路径
                string byzeDirectory = Path.Combine(userDirectory, "Byze");

                // 如果 Byze 文件夹不存在，则创建
                if (!Directory.Exists(byzeDirectory))
                {
                    Directory.CreateDirectory(byzeDirectory);
                }

                // 构造 .byze 文件路径
                string byzeFilePath = Path.Combine(byzeDirectory, ".byze");

                // 将 JSON 写入 .byze 文件
                await File.WriteAllTextAsync(byzeFilePath, config);

                // 返回文件路径
                return byzeFilePath;
            }
            catch (Exception ex)
            {
                throw new Exception($"导出配置文件失败: {ex.Message}");
            }
        }

        // Chat
        public async Task<string> ChatAsync(object data, bool isStream = false, Action<JsonElement>? onData = null, Action<string>? onError = null, Action? onEnd = null)
        {
            try
            {
                var json = JsonSerializer.Serialize(data);
                var content = new StringContent(json, System.Text.Encoding.UTF8, "application/json");

                if (isStream)
                {
                    var request = new HttpRequestMessage(HttpMethod.Post, "/services/chat")
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

                    return "Stream completed";
                }
                else
                {
                    var response = await _client.PostAsync("/services/chat", content);
                    response.EnsureSuccessStatusCode();

                    return await response.Content.ReadAsStringAsync();
                }
            }
            catch (Exception ex)
            {
                throw new Exception($"Chat 服务请求失败: {ex.Message}");
            }
        }

        // Generate
        public async Task<string> GenerateAsync(object data, bool isStream = false, Action<JsonElement>? onData = null, Action<string>? onError = null, Action? onEnd = null)
        {
            try
            {
                var json = JsonSerializer.Serialize(data);
                var content = new StringContent(json, System.Text.Encoding.UTF8, "application/json");

                if (isStream)
                {
                    var request = new HttpRequestMessage(HttpMethod.Post, "/services/generate")
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

                    return "Stream completed";
                }
                else
                {
                    var response = await _client.PostAsync("/services/generate", content);
                    response.EnsureSuccessStatusCode();

                    return await response.Content.ReadAsStringAsync();
                }
            }
            catch (Exception ex)
            {
                throw new Exception($"Generate 服务请求失败: {ex.Message}");
            }
        }

        // embed
        public async Task<string> EmbedAsync(object data)
        {
            return await RequestAsync(HttpMethod.Post, "/services/embed", data);
        }

        // text-to-image
        public async Task<string> TextToImageAsync(object data)
        {
            return await RequestAsync(HttpMethod.Post, "/services/text-to-image", data);
        }

        // 检查 byze 状态
        public async Task<bool> IsByzeAvailiableAsync()
        {
            try
            {
                var response = await _client.GetAsync("/");
                return response.IsSuccessStatusCode;
            }
            catch (Exception ex)
            {
                throw new Exception($"检查 Byze 状态失败: {ex.Message}");
            }
        }

        // 检查 byze 是否下载
        public bool IsByzeExisted()
        {
            try
            {
                // 获取用户目录
                string userDirectory = Environment.GetFolderPath(Environment.SpecialFolder.UserProfile);

                // 根据操作系统设置路径
                string byzePath;
                if (OperatingSystem.IsWindows())
                {
                    byzePath = Path.Combine(userDirectory, "Byze", "byze.exe");
                }
                else if (OperatingSystem.IsMacOS())
                {
                    byzePath = Path.Combine(userDirectory, "Byze", "byze");
                }
                else
                {
                    throw new PlatformNotSupportedException("当前操作系统不支持");
                }

                // 检查文件是否存在
                return File.Exists(byzePath);
            }
            catch (Exception ex)
            {
                Console.WriteLine($"检查 Byze 文件失败: {ex.Message}");
                return false;
            }
        }

        



    }
}
