
const getServicesSchema = {
    type: "object",
    properties: {
        business_code: { type: "integer" },
        message: { type: "string" },
        data: {
            type: "array",
            items: {
                type: "object",
                properties: {
                    service_name: { type: "string" },
                    hybrid_policy: { type: "string" },
                    remote_provider: { type: "string" },
                    local_provider: { type: "string" },
                    status: { type: "integer" },
                    created_at: { type: "string", format: "date-time" },
                    updated_at: { type: "string", format: "date-time" }
                },
                required: [
                    "service_name",
                    "hybrid_policy",
                    "remote_provider",
                    "local_provider",
                    "status",
                ]
            }
        }
    },
    required: ["business_code", "message", "data"]
};

const installServiceRequestSchema = {
    type: "object",
    properties: {
        service_name: { type: "string" },
        service_source: { type: "string", enum: ["remote", "local"] },
        hybrid_policy: { type: "string" },
        flavor_name: { type: "string" },
        provider_name: { type: "string" },
        auth_type: { type: "string", enum: ["apikey", "oauth", "none"] },
        auth_key: { type: "string" }
    },
    required: [
        "service_name",
        "service_source",
        "hybrid_policy",
        "flavor_name",
        "provider_name",
    ]
};

const updateServiceRequestSchema = {
    type: "object",
    properties: {
        service_name: { type: "string" },
        hybrid_policy: { type: "string" },
        remote_provider: { type: "string" },
        local_provider: { type: "string" },
    },
    required: [
        "service_name"
    ]
};





const ResponseSchema = {
    type: "object",
    properties: {
        business_code: { type: "integer" },
        message: { type: "string" }
    },
    required: ["business_code", "message"]
};

module.exports = {
    getServicesSchema,
    installServiceRequestSchema,
    ResponseSchema,
    updateServiceRequestSchema
};
    
