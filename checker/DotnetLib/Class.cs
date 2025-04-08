namespace ByzeLib;
using System;
using System.Net.Http;
using System.Text;
using System.Text.Json;
using System.Threading.Tasks;
using System.Diagnostics;
using System.IO;
using System.Linq;
using System.Text.Json.Nodes;
using 

public class Byze
{
    private readonly HttpClient _client;
    private readonly string _version = "byze/v0.2";

    public Byze(string version = null)
    {
        if (!string.IsNullOrEmpty(version))
        {
            _version = version;
        }

        _client = new HttpClient
        {
            BaseAddress = new Uri($"http://localhost:16688/{_version}")
        };
        _client.DefaultRequestHeaders.Accept.Clear();
        _client.DefaultRequestHeaders.Accept.Add(new System.Net.Http.Headers.MediaTypeWithQualityHeaderValue("application/json"));
    }

    private async Task<T> ValidateSchema<T>(string schemaJson, T data)
    {
        if (data == null || (data is JsonNode node && node.AsObject().Count == 0))
        {
            // 如果 data 为空或是空对象，跳过验证
            return data;
        }

        try
        {
            JsonSchema schema = await JsonSchema.FromJsonAsync(schemaJson);
            JsonNode jsonData = JsonSerializer.SerializeToNode(data);
            ValidationResults results = schema.Validate(jsonData);

            if (results.IsValid)
            {
                return data;
            }
            else
            {
                throw new Exception($"Schema validation failed: {JsonSerializer.Serialize(results.Errors)}");
            }
        }
        catch (Exception ex)
        {
            throw new Exception($"Schema validation error: {ex.Message}");
        }
    }



    // 使用 ValidateSchema 函数进行验证
    public async Task<Service[]> GetServices()
    {
        HttpResponseMessage response = await _client.GetAsync("/service");
        response.EnsureSuccessStatusCode();
        string responseBody = await response.Content.ReadAsStringAsync();
        Service[] services = JsonSerializer.Deserialize<Service[]>(responseBody);
        return await ValidateSchema(GetServicesSchemaJson, services);
    }

    public async Task<Response> InstallService(Service data)
    {
        string json = JsonSerializer.Serialize(data);
        StringContent content = new StringContent(json, Encoding.UTF8, "application/json");
        HttpResponseMessage response = await _client.PostAsync("/service", content);
        response.EnsureSuccessStatusCode();
        string responseBody = await response.Content.ReadAsStringAsync();
        Response result = JsonSerializer.Deserialize<Response>(responseBody);
        return await ValidateSchema(InstallServiceRequestSchemaJson, result);
    }

    public async Task<Response> UpdateService(Service data)
    {
        string json = JsonSerializer.Serialize(data);
        StringContent content = new StringContent(json, Encoding.UTF8, "application/json");
        HttpResponseMessage response = await _client.PutAsync("/service", content);
        response.EnsureSuccessStatusCode();
        string responseBody = await response.Content.ReadAsStringAsync();
        Response result = JsonSerializer.Deserialize<Response>(responseBody);
        return await ValidateSchema(UpdateServiceRequestSchemaJson, result);
    }

    // ... 其他方法
}

// 示例数据模型类
public class Service
{
    // ... 服务属性
}

public class Response
{
    // ... 响应属性
}
