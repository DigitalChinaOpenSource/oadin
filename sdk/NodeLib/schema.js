//*****************************************************************************
// Copyright 2025 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//*****************************************************************************

const { str } = require("ajv");

// ========== 服务管理 ==========
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
        auth_type: { type: "string", enum: ["apikey", "none", "credentials"] },
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

// ========== 模型管理 ==========
const getModelsSchema = {
    type: "object",
    properties: {
        business_code: { type: "integer" },
        message: { type: "string" },
        data: {
            type: "array",
            items: {
                type: "object",
                properties: {
                    model_name: { type: "string" },
                    povider_name: { type: "string" },
                    status: { type: "string" },
                    created_at: { type: "string", format: "date-time" },
                    updated_at: { type: "string", format: "date-time" }
                },
                required: [
                    "model_name",
                    "provider_name",
                    "status",
                ]
            }
        }
    },
    required: ["business_code", "message", "data"]
};

const installModelRequestSchema = {
    type: "object",
    properties: {
        model_name: { type: "string" },
        service_name: { type: "string" },
        service_source: { type: "string"},
        provider_name: { type: "string" },
    },
    required: [
        "model_name",
        "service_name",
        "service_source",
    ]
};

const deleteModelRequestSchema = {
    type: "object",
    properties: {
        model_name: { type: "string" },
        service_name: { type: "string" },
        service_source: { type: "string"},
        provider_name: { type: "string" },
    },
    required: [
        "model_name",
        "service_name",
        "service_source",
    ]
};

const installModelStreamResponseSchema = {
    type: "object",
    properties: {
        status: { type: "string" },
        digest: { type: "string" },
        total: { type: "integer" },
        completed: { type: "integer" }
    },
    required: ["status"]
};

const cancelModelStreamRequestSchema = {
    type: "object",
    properties: {
        model_name: { type: "string" }
    },
    required: ["model_name"]
};

// ========== 服务提供商管理 ==========
const getServiceProvidersSchema = {
    type: "object",
    properties: {
        business_code: { type: "integer" },
        message: { type: "string" },
        data: {
            type: "array",
            items: {
                type: "object",
                properties: {
                    provider_name: { type: "string" },
                    service_name: { type: "string" },
                    service_source: { type: "string" },
                    desc: { type: "string" },
                    auth_type: { type: "string", enum: ["apikey", "none", "credentials"] },
                    auth_key: { type: "string" },
                    flavor: { type: "string" },
                    properties: { type: "string" },
                    models: { type: ["array", "null"], items: { type: "string" } },
                    status: { type: "integer" },
                    created_at: { type: "string", format: "date-time" },
                    updated_at: { type: "string", format: "date-time" }
                },
                required: [
                    "provider_name",
                    "status",
                ]
            }
        }
    },
    required: ["business_code", "message", "data"]
};

const installServiceProviderRequestSchema = {
    type: "object",
    properties: {
        service_name: { type: "string" },
        service_source: { type: "string" },
        flavor_name: { type: "string" },
        provider_name: { type: "string" },
        desc: { type: "string" },
        method: { type: "string" },
        auth_type: { type: "string", enum: ["apikey", "none", "credentials"] },
        auth_key: { type: "string" },
        models: { type: "array", items: { type: "string" } },
        extra_headers: { type: "object" },
        extra_json_body: { type: "object" },
        properties: { type: "object" },
    },
    required: [
        "service_name",
        "service_source",
        "flavor_name",
        "provider_name",
    ]
};

const updateServiceProviderRequestSchema = {
    type: "object",
    properties: {
        service_name: { type: "string" },
        service_source: { type: "string" },
        flavor_name: { type: "string" },
        provider_name: { type: "string" },
        desc: { type: "string" },
        method: { type: "string" },
        auth_type: { type: "string", enum: ["apikey", "none", "credentials"] },
        auth_key: { type: "string" },
        models: { type: "array", items: { type: "string" } },
        extra_headers: { type: "object" },
        extra_json_body: { type: "object" },
        properties: { type: "object" },
    },
    required: [
        "service_name",
        "service_source",
        "flavor_name",
        "provider_name",
    ]
};

const deleteServiceProviderRequestSchema = {
    type: "object",
    properties: {
        provider_name: { type: "string" },
    },
    required: [
        "provider_name",
    ]   
};

// ========= 导入导出 ==========
const exportRequestSchema = {
    type: "object",
    properties: {
        service_name: { type: "string" }
    },
};

const jsonConfigSchema = {
    type: "object",
    properties: {
        version: { type: "string" },
        chat: {
            type: "object",
            properties: {
                hybrid_policy: { type: "string" },
                services: {
                    type: "object",
                    properties: {
                        local: {
                            $ref: "#/definitions/serviceSchema"
                        },
                        remote: {
                            $ref: "#/definitions/serviceSchema"
                        }
                    }
                }
            },
            required: ["hybrid_policy", "services"]
        },
        embed: {
            type: "object",
            properties: {
                hybrid_policy: { type: "string" },
                status: { type: "string" },
                services: {
                    type: "object",
                    properties: {
                        local: { type: "object" },
                        remote: { type: "object" }
                    }
                }
            },
            required: ["hybrid_policy", "status", "services"]
        }
    },
    required: ["chat", "embed"],
    definitions: {
        serviceSchema: {
            type: "object",
            properties: {
                service_id: { type: "integer" },
                service_type: { type: "string", enum: ["local", "remote"] },
                provider: {
                    type: "object",
                    properties: {
                        name: { type: "string" },
                        desc: { type: "string" },
                        method: { type: "string", enum: ["POST", "GET"] },
                        url: { type: "string", format: "uri" },
                        api_flavor: { type: "string" },
                        properties: {
                            type: "object",
                            properties: {
                                supported_response_mode: {
                                    type: "array",
                                    items: { type: "string", enum: ["sync", "stream"] }
                                },
                                models: {
                                    type: "array",
                                    items: { type: "string" }
                                }
                            },
                            required: ["supported_response_mode", "models"]
                        },
                        status: { type: "string" }
                    },
                    required: ["name", "method", "url", "api_flavor", "properties", "status"]
                },
                created_at: { type: "string", format: "date-time" },
                updated_at: { type: "string", format: "date-time" }
            },
            required: ["service_type", "provider"]
        }
    }
};



// ========= 以上的响应 ==========
const ResponseSchema = {
    type: "object",
    properties: {
        business_code: { type: "integer" },
        message: { type: "string" }
    },
    required: ["business_code", "message"]
};

// ========= 获取模型/支持模型/推荐模型 ==========
const modelsResponse = {
    type: "object",
    properties: {
        models: {
            type: "array",
            items: {
                type: "object",
                properties: {
                    digest: { type: "string" },
                    family: { type: "string" },
                    format: { type: "string" },
                    modified_at: { type: "string", format: "date-time" },
                    name: { type: "string" },
                    parameter_size: { type: "string" },
                    quantization_level: { type: "string" },
                    size: { type: "integer" }
                },
                required: [
                    "name",
                ]
            }
        }
    },
    required: ["models"]
}

const getModelsSupported = {
    type: "object",
    properties: {
        service_source: { type: "string", enum: ["local", "remote"]},
        flavor: { type:"string" }
    },
    required: [ "service_source", "flavor"]
}

const recommendModelsResponse = {
    type: "object",
    properties: {
        business_code: { type: "integer" },
        message: { type: "string" },
        data: {
            type: "object",
            properties: {
                chat: {
                    type: "array",
                    items: {
                        type: "object",
                        properties: {
                            service_name: { type: "string" },
                            api_flavor: { type: "string" },
                            method: { type: "string" },
                            desc: { type: "string" },
                            url: { type: "string" },
                            auth_type: { type: "string" },
                            auth_apply_url: { type: "string" },
                            auth_fields: { type: ["array", "null"], items: { type: "string" } },
                            name: { type: "string" },
                            service_provider_name: { type: "string" },
                            size: { type: "string" },
                            is_recommended: { type: "boolean" }
                        },
                        required: ["name", "size", "is_recommended"]
                    }
                }
            },
            required: ["chat"]
        }
    },
    required: ["business_code", "message", "data"]
};

const SmartvisionModelSupportRequest = {
    type: "object",
    properties: {
        env_type: { type: "string", enum: ["dev", "product"] }
    },
    required: ["env_type"]
};

const SmartvisionModelSupport = {
    type: "object",
    properties: {
        business_code: { type: "integer" },
        message: { type: "string" },
        data: {
            type: "object",
            properties: {
                code: { type: "integer" },
                data: {
                    type: "array",
                    items: {
                        type: "object",
                        properties: {
                            id: { type: "integer" },
                            name: { type: "string" },
                            avatar: { type: "string" },
                            type: { type: "integer" },
                            provider: { type: "string" },
                            modelKey: { type: "string" },
                            credentialParamsId: { type: "string" },
                            introduce: { type: "string" },
                            tags: {
                                type: "array",
                                items: { type: "string" }
                            },
                            credentialParams: {
                                type: "array",
                                items: {
                                    type: "object",
                                    properties: {
                                        id: { type: "integer" },
                                        name: { type: "string" },
                                        label: { type: "string" },
                                        type: { type: "string" },
                                        placeholder: { type: "string" },
                                        required: { type: "integer" },
                                        value: { type: ["string", "null"] },
                                        sort: { type: "integer" },
                                        createTime: { type: "integer" },
                                        updateTime: { type: "integer" }
                                    } 
                                }
                            }
                        }
                    }
                }
            }
        }
    }
};

// ========= chat ==========
const chatRequest = {
    type: "object",
    properties: {
        model: { type: "string" },
        stream: { type: "boolean" },
        messages: {
            type: "array",
            items: {
                type: "object",
                properties: {
                    role: { type: "string", enum: ["user", "assistant", "system"] },
                    content: { type: "string" }
                },
                required: ["role", "content"]
            }
        },
        options:{
            type: "object",
            properties: {
                temperature: { type: "number" },
                max_tokens: { type: "integer" },
            }
        }
    },
    required: ["model", "messages"]
};

const chatResponse = {
    type: "object",
    properties: {
        created_at: { type: "string", format: "date-time" },
        finish_reason: { type: "string" },
        finished: { type: "boolean" },
        id: { type: "string" },
        message: {
            type: "object",
            properties: {
                content: { type: "string" },
                role: { type: "string", enum: ["assistant"] }
            },
            required: ["content", "role"]
        },
        model: { type: "string" }
    },
    required: ["message", "model"]
};


// ========= generate ==========
const generateRequest = {
    type: "object",
    properties: {
        model: { type: "string" },
        stream: { type: "boolean" },
        prompt: { type: "string" }
    },
    required: ["model", "prompt"]
};  

const generateResponse = {
    type: "object",
    properties: {
        model: { type: "string" },
        created_at: { type: "string", format: "date-time" },
        message: {
            type: "object",
            properties: {
                role: { type: "string", enum: ["assistant"] },
                content: { type: "string" }
            },
            required: ["role", "content"]
        },
        done: { type: "boolean" },
        done_reason: { type: "string" },
        total_duration: { type: "integer" },
        load_duration: { type: "integer" },
        prompt_eval_count: { type: "integer" },
        prompt_eval_duration: { type: "integer" },
        eval_count: { type: "integer" },
        eval_duration: { type: "integer" }
    },
    required: ["model", "created_at", "message"]
}

// ========= text-to-image ==========
const textToImageRequest = {
    type: "object",
    properties: {
        model: { type: "string" },
        prompt: { type: "string" }
    },
    required: ["prompt"]
};

const textToImageResponse = {
    type: "object",
    properties: {
        data: {
            type: "object",
            properties: {
                url: { type: "string" }
            },
        },
        id: { type: "string" },
    },
    required: ["data"]
};

// ========= embedding ==========
const embeddingRequest = {
    type: "object",
    properties: {
        model: { type: "string" },
        input: { 
            type: "array",
            items: {
                type: "string"
            }
        }
    },
    required: ["model", "input"]
};

const embeddingResponse = {
    type: "object",
    properties: {
        model: { type: "string" },
        id: { type: "string" },
        data: {
            type: "array",
            items: {
                type: "object",
                properties: {
                    embedding: {
                        type: "array",
                        items: { type: "number" }
                    },
                    index: { type: "integer" },
                    object: { type: "string" },
                },
                required: ["embedding", "index"]
            }
        }
    },
    required: ["model", "data"]
};

const speechToTextRequest = {
    type: "object",
    properties: {
        model: { type: "string" },
        audio: { type: "string" },
        language: { type: "string" }
    }
};

const speechToTextResponse = {
    type: "object",
    properties: {
        business_code: { type: "integer" },
        message: { type: "string" },
        data: { 
            type: "object",
            properties: {
                data: {
                    type: "object",
                    properties: {
                        segments: {
                            type: "array",
                            items: {
                                type: "object",
                                properties: {
                                    end: { type: "string" },
                                    id: { type: "integer"},
                                    start: { type: "string"},
                                    text: { type: "string"}
                                }
                            }
                        }
                    }
                }
            }
        }
    }
};

module.exports = {
    getServicesSchema,
    installServiceRequestSchema,
    ResponseSchema,
    updateServiceRequestSchema,
    getModelsSchema,
    installModelRequestSchema,
    deleteModelRequestSchema,
    getServiceProvidersSchema,
    installModelStreamResponseSchema,
    cancelModelStreamRequestSchema,
    installServiceProviderRequestSchema,
    updateServiceProviderRequestSchema,
    deleteServiceProviderRequestSchema,
    exportRequestSchema,
    jsonConfigSchema,
    modelsResponse,
    getModelsSupported,
    recommendModelsResponse,
    SmartvisionModelSupportRequest,
    SmartvisionModelSupport,
    chatRequest,
    chatResponse,
    generateRequest,
    generateResponse,
    textToImageRequest,
    textToImageResponse,
    embeddingRequest,
    embeddingResponse,
    speechToTextRequest,
    speechToTextResponse
};
    
