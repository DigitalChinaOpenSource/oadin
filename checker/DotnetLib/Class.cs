using System;
using System.Web;
using System.Diagnostics;
using System.Threading;
using System.Threading.Tasks;
using Microsoft.Win32;
using System.Net.Http;
using System.Text.Json;
using System.Threading.Tasks;

namespace Oadin
{
    public class OadinClient
    {
        private readonly HttpClient _client;
        private readonly string _baseUrl;

        public OadinClient(string version = "oadin/v0.2")
        {
            if (!version.EndsWith("/")) version += "/";
            _baseUrl = $"http://127.0.0.1:16688/{version}";
            _client = new HttpClient { BaseAddress = new Uri(_baseUrl) };
        }
        
        // 通用请求方法
        private async Task<string> RequestAsync(HttpMethod method, string path, object? data = null)
        {
            try
            {
                if (path.StartsWith("/"))
                {
                    path = path.TrimStart('/');
                }

                HttpRequestMessage request = new HttpRequestMessage(method, path); // 使用相对路径

                if (data != null)
                {
                    var json = JsonSerializer.Serialize(data);
                    request.Content = new StringContent(json, System.Text.Encoding.UTF8, "application/json");
                }
                Console.WriteLine($"Request URL: {request.RequestUri}");
                Console.WriteLine($"Headers: {string.Join(", ", request.Headers)}");

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
                var request = new HttpRequestMessage(HttpMethod.Post, $"{_baseUrl}model/stream")
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
        public async Task<string> GetSmartvisionModelsSupportedAsync(Dictionary<string, string> headers)
        {
            // 构建带查询参数的 URL
            var queryParams = HttpUtility.ParseQueryString(string.Empty);
            foreach (var header in headers)
            {
                if (header.Key.ToLower() == "env_type") // 忽略大小写
                {
                    queryParams[header.Key] = header.Value;
                }
            }
            string fullPath = $"/model/support/smartvision?{queryParams.ToString()}";

            return await RequestAsync(HttpMethod.Get, fullPath, null); // 不再传递 headers
        }

        // 导入配置文件
        public async Task<string> ImportConfigAsync(string filePath)
        {
            // 读取json文件内容
            string jsonContent = await File.ReadAllTextAsync(filePath);
            var jsonElement = JsonSerializer.Deserialize<JsonElement>(jsonContent);
            return await RequestAsync(HttpMethod.Post, "/config/import", jsonElement);
        }

        // 导出配置文件
        public async Task<string> ExportConfigAsync(object data)
        {
            try
            {
                // 调用 RequestAsync 获取配置文件的 JSON 响应
                var config = await RequestAsync(HttpMethod.Get, "/config/export", data);

                string userDirectory = Environment.GetFolderPath(Environment.SpecialFolder.UserProfile);
                string oadinDirectory = Path.Combine(userDirectory, "Oadin");
                string oadinFilePath = Path.Combine(oadinDirectory, ".oadin");

                await File.WriteAllTextAsync(oadinFilePath, config);

                // 返回文件路径
                return oadinFilePath;
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
                    var request = new HttpRequestMessage(HttpMethod.Post, $"{_baseUrl}services/chat")
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

                            if (responseData.TryGetProperty("finished", out var finishedProperty) && finishedProperty.GetBoolean())
                            {
                                onEnd?.Invoke();
                                break;
                            }
                        }
                        catch (JsonException jsonEx)
                        {
                            onError?.Invoke($"JSON 解析错误: {jsonEx.Message}");    
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
                    var response = await _client.PostAsync("services/chat", content);
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
                    var request = new HttpRequestMessage(HttpMethod.Post, $"{_baseUrl}services/generate")
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

        // 检查 oadin 状态
        public async Task<bool> IsOadinAvailiableAsync()
        {
            try
            {
                var response = await _client.GetAsync("/");
                return response.IsSuccessStatusCode;
            }
            catch (Exception ex)
            {
                throw new Exception($"检查 Oadin 状态失败: {ex.Message}");
            }
        }

        // 检查 oadin 是否下载
        public bool IsOadinExisted()
        {
            try
            {
                // 获取用户目录
                string userDirectory = Environment.GetFolderPath(Environment.SpecialFolder.UserProfile);

                // 根据操作系统设置路径
                string oadinPath;
                if (OperatingSystem.IsWindows())
                {
                    oadinPath = Path.Combine(userDirectory, "Oadin", "oadin.exe");
                }
                else if (OperatingSystem.IsMacOS())
                {
                    oadinPath = Path.Combine(userDirectory, "Oadin", "oadin");
                }
                else
                {
                    throw new PlatformNotSupportedException("当前操作系统不支持");
                }

                // 检查文件是否存在
                return File.Exists(oadinPath);
            }
            catch (Exception ex)
            {
                Console.WriteLine($"检查 Oadin 文件失败: {ex.Message}");
                return false;
            }
        }

        // 下载 Oadin
        public async Task<bool> DownloadOadinAsync()
        {
            try
            {
                // 根据操作系统选择下载 URL 和目标路径
                string url = OperatingSystem.IsMacOS()
                    ? "http://120.232.136.73:31619/oadindev/oadin.zip"
                    : "http://120.232.136.73:31619/oadindev/oadin.exe";

                string userDirectory = Environment.GetFolderPath(Environment.SpecialFolder.UserProfile);
                string oadinDirectory = Path.Combine(userDirectory, "Oadin");
                string destFileName = OperatingSystem.IsMacOS() ? "oadin.zip" : "oadin.exe";
                string destFilePath = Path.Combine(oadinDirectory, destFileName);

                // 创建 Oadin 目录
                if (!Directory.Exists(oadinDirectory))
                {
                    Directory.CreateDirectory(oadinDirectory);
                }

                // 下载文件
                using var client = new HttpClient();
                using var response = await client.GetAsync(url, HttpCompletionOption.ResponseHeadersRead);
                response.EnsureSuccessStatusCode();

                using var fileStream = new FileStream(destFilePath, FileMode.Create, FileAccess.Write, FileShare.None);
                await response.Content.CopyToAsync(fileStream);

                Console.WriteLine($"✅ 下载完成: {destFilePath}");

                // 如果是 macOS，解压 ZIP 文件
                if (OperatingSystem.IsMacOS())
                {
                    string extractedPath = Path.Combine(oadinDirectory, "oadin");
                    System.IO.Compression.ZipFile.ExtractToDirectory(destFilePath, oadinDirectory, true);
                    File.Delete(destFilePath);
                    Console.WriteLine($"✅ 解压完成: {extractedPath}");

                    // 设置可执行权限
                    if (File.Exists(extractedPath))
                    {
                        var chmod = new ProcessStartInfo
                        {
                            FileName = "chmod",
                            Arguments = "+x " + extractedPath,
                            RedirectStandardOutput = true,
                            RedirectStandardError = true,
                            UseShellExecute = false,
                            CreateNoWindow = true
                        };
                        Process.Start(chmod)?.WaitForExit();
                    }
                }

                // 添加 Oadin 目录到环境变量
                AddToEnvironmentPath(oadinDirectory);

                return true;
            }
            catch (Exception ex)
            {
                Console.WriteLine($"❌ 下载 Oadin 失败: {ex.Message}");
                return false;
            }
        }

        // 将路径添加到环境变量
        private void AddToEnvironmentPath(string directory)
        {
            try
            {
                if (OperatingSystem.IsWindows())
                {
                    // Windows: 修改注册表以永久添加到 PATH
                    const string regKey = @"Environment";
                    using var key = Registry.CurrentUser.OpenSubKey(regKey, writable: true);
                    if (key == null) throw new Exception("无法打开注册表键");

                    string? currentPath = key.GetValue("Path", "", RegistryValueOptions.DoNotExpandEnvironmentNames)?.ToString();
                    if (currentPath == null || !currentPath.Contains(directory))
                    {
                        string newPath = string.IsNullOrEmpty(currentPath) ? directory : $"{currentPath};{directory}";
                        key.SetValue("Path", newPath, RegistryValueKind.ExpandString);
                        Console.WriteLine("✅ 已将 Oadin 目录添加到环境变量 PATH");
                    }
                    else
                    {
                        Console.WriteLine("✅ Oadin 目录已存在于环境变量 PATH 中");
                    }
                }
                else if (OperatingSystem.IsMacOS() || OperatingSystem.IsLinux())
                {
                    // macOS/Linux: 修改 shell 配置文件
                    string shellConfigPath = Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.UserProfile), ".zshrc");
                    if (!File.Exists(shellConfigPath))
                    {
                        shellConfigPath = Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.UserProfile), ".bashrc");
                    }

                    string exportLine = $"export PATH=\"$PATH:{directory}\"";
                    if (File.Exists(shellConfigPath))
                    {
                        string content = File.ReadAllText(shellConfigPath);
                        if (!content.Contains(exportLine))
                        {
                            File.AppendAllText(shellConfigPath, Environment.NewLine + exportLine);
                            Console.WriteLine($"✅ 已将 Oadin 目录添加到 {Path.GetFileName(shellConfigPath)}，请执行以下命令使其生效：\nsource {shellConfigPath}");
                        }
                        else
                        {
                            Console.WriteLine($"✅ Oadin 目录已存在于 {Path.GetFileName(shellConfigPath)} 中");
                        }
                    }
                    else
                    {
                        File.WriteAllText(shellConfigPath, exportLine + Environment.NewLine);
                        Console.WriteLine($"✅ 已创建 {Path.GetFileName(shellConfigPath)} 并添加 Oadin 目录，请执行以下命令使其生效：\nsource {shellConfigPath}");
                    }
                }
                else
                {
                    throw new PlatformNotSupportedException("当前操作系统不支持添加环境变量");
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine($"❌ 添加 Oadin 目录到环境变量失败: {ex.Message}");
            }
        }

        // 启动 Oadin 服务
        public bool InstallOadin()
        {
            try
            {
                string userDirectory = Environment.GetFolderPath(Environment.SpecialFolder.UserProfile);
                string oadinDirectory = Path.Combine(userDirectory, "Oadin");
                string oadinExecutable = OperatingSystem.IsMacOS()
                    ? Path.Combine(oadinDirectory, "oadin")
                    : Path.Combine(oadinDirectory, "oadin.exe");

                if (!File.Exists(oadinExecutable))
                {
                    Console.WriteLine("❌ Oadin 可执行文件不存在，请先下载。");
                    return false;
                }

                // 确保 PATH 包含 Oadin 目录
                string pathEnv = Environment.GetEnvironmentVariable("PATH") ?? string.Empty;
                if (!pathEnv.Contains(oadinDirectory))
                {
                    Environment.SetEnvironmentVariable("PATH", pathEnv + Path.PathSeparator + oadinDirectory);
                }

                // 启动 Oadin 服务
                var processStartInfo = new ProcessStartInfo
                {
                    FileName = oadinExecutable,
                    Arguments = "server start -d",
                    RedirectStandardOutput = true,
                    RedirectStandardError = true,
                    UseShellExecute = false,
                    CreateNoWindow = true
                };

                using var process = Process.Start(processStartInfo);
                if (process == null)
                {
                    Console.WriteLine("❌ 启动 Oadin 服务失败。");
                    return false;
                }

                process.WaitForExit();

                if (process.ExitCode == 0)
                {
                    Console.WriteLine("✅ Oadin 服务已启动。");
                    return true;
                }
                else
                {
                    Console.WriteLine($"❌ Oadin 服务启动失败，退出码: {process.ExitCode}");
                    return false;
                }
            }
            catch (Exception ex)
            {
                Console.WriteLine($"❌ 启动 Oadin 服务失败: {ex.Message}");
                return false;
            }
        }


    }
}
