public class ModelData
{
    public string ModelName { get; set; }
    public string ProviderName { get; set; }
    public string Status { get; set; }
    public DateTime CreatedAt { get; set; }
    public DateTime UpdatedAt { get; set; }
}

public class GetModelsResponse
{
    public int BusinessCode { get; set; }
    public string Message { get; set; }
    public List<ModelData> Data { get; set; }
}