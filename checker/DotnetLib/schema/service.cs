public class GetServicesResponse
{
    public int BusinessCode { get; set; }
    public string Message { get; set; }
    public List<Service> Data { get; set; }
}

public class Service
{
    public string ServiceName { get; set; }
    public string HybridPolicy { get; set; }
    public string RemoteProvider { get; set; }
    public string LocalProvider { get; set; }
    public int Status { get; set; }
    public DateTime CreatedAt { get; set; }
    public DateTime UpdatedAt { get; set; }
}

public class InstallServiceRequest
{
    public string ServiceName { get; set; }
    public string ServiceSource { get; set; }
    public string HybridPolicy { get; set; }
    public string FlavorName { get; set; }
    public string ProviderName { get; set; }
    public string AuthType { get; set; }
    public string AuthKey { get; set; }
}

public class UpdateServiceRequest
{
    public string ServiceName { get; set; }
    public string HybridPolicy { get; set; }
    public string RemoteProvider { get; set; }
    public string LocalProvider { get; set; }
}

public class DeleteServiceRequest
{
    public string ServiceName { get; set; }
    public string ServiceSource { get; set; }
}
