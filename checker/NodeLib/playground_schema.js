const baseResponseSchema = {
    type: "object",
    properties: {
        bcode: {
            type: "object",
            properties: {
                code: { type: "integer" },
                message: { type: "string" }
            },
            required: ["code"]
        }
    },
    required: ["bcode"]
};

// 创建会话请求
const createSessionRequestSchema = {
    type: "object",
    properties: {
        title: { type: "string" },
        modelId: { type: "string" },
        embedModelId: { type: "string" }
    },
    required: ["modelId"]
};

// 创建会话响应
const createSessionResponseSchema = {
    type: "object",
    properties: {
        bcode: {
            type: "object",
            properties: {
                code: { type: "integer" },
                message: { type: "string" }
            },
            required: ["code"]
        },
        data: {
            type: "object",
            properties: {
                id: { type: "string" },
                title: { type: "string" },
                modelId: { type: "string" },
                modelName: { type: "string" },
                embedModelId: { type: "string" },
                thinkingEnabled: { type: "boolean" },
                createdAt: { type: "string" },
                updatedAt: { type: "string" }
            },
            required: ["id", "modelId", "modelName"]
        }
    },
    required: ["bcode", "data"]
};

// 获取会话列表响应
const getSessionsResponseSchema = {
    type: "object",
    properties: {
        bcode: {
            type: "object",
            properties: {
                code: { type: "integer" },
                message: { type: "string" }
            },
            required: ["code"]
        },
        data: {
            type: "array",
            items: {
                type: "object",
                properties: {
                    id: { type: "string" },
                    title: { type: "string" },
                    modelId: { type: "string" },
                    modelName: { type: "string" },
                    embedModelId: { type: "string" },
                    thinkingEnabled: { type: "boolean" },
                    createdAt: { type: "string" },
                    updatedAt: { type: "string" }
                },
                required: ["id", "modelId", "modelName"]
            }
        }
    },
    required: ["bcode", "data"]
};

// 发送消息请求
const sendMessageRequestSchema = {
    type: "object",
    properties: {
        session_id: { type: "string" },
        content: { type: "string" }
    },
    required: ["session_id", "content"]
};

// 发送消息响应
const sendMessageResponseSchema = {
    type: "object",
    properties: {
        bcode: {
            type: "object",
            properties: {
                code: { type: "integer" },
                message: { type: "string" }
            },
            required: ["code"]
        },
        data: {
            type: "object",
            properties: {
                id: { type: "string" },
                session_id: { type: "string" },
                role: { type: "string" },
                content: { type: "string" },
                modelId: { type: "string" },
                modelName: { type: "string" },
                createdAt: { type: "string" },
                thoughts: { type: "string" }
            },
            required: ["id", "session_id", "role", "content", "modelId", "modelName"]
        }
    },
    required: ["bcode", "data"]
};

// 发送流式消息请求
const sendStreamMessageRequestSchema = {
    type: "object",
    properties: {
        session_id: { type: "string" },
        content: { type: "string" }
    },
    required: ["session_id", "content"]
};

// 获取消息历史请求
const getMessagesRequestSchema = {
    type: "object",
    properties: {
        session_id: { type: "string" }
    },
    required: ["session_id"]
};

// 获取消息历史响应
const getMessagesResponseSchema = {
    type: "object",
    properties: {
        bcode: {
            type: "object",
            properties: {
                code: { type: "integer" },
                message: { type: "string" }
            },
            required: ["code"]
        },
        data: {
            type: "array",
            items: {
                type: "object",
                properties: {
                    id: { type: "string" },
                    sessionId: { type: "string" },
                    role: { type: "string" },
                    content: { type: "string" },
                    modelId: { type: "string" },
                    modelName: { type: "string" },
                    createdAt: { type: "string" },
                    thoughts: { type: "string" }
                },
                required: ["id", "sessionId", "role", "content", "modelId", "modelName"]
            }
        }
    },
    required: ["bcode", "data"]
};

// 上传文件响应
const uploadFileResponseSchema = {
    type: "object",
    properties: {
        bcode: {
            type: "object",
            properties: {
                code: { type: "integer" },
                message: { type: "string" }
            },
            required: ["code"]
        },
        data: {
            type: "object",
            properties: {
                id: { type: "string" },
                sessionId: { type: "string" },
                name: { type: "string" },
                path: { type: "string" },
                size: { type: "integer" },
                type: { type: "string" },
                createdAt: { type: "string" },
                updatedAt: { type: "string" }
            },
            required: ["id", "sessionId", "name"]
        }
    },
    required: ["bcode", "data"]
};

// 处理文件请求
const processFileRequestSchema = {
    type: "object",
    properties: {
        fileId: { type: "string" },
        model: { type: "string" }
    },
    required: ["fileId", "model"]
};

// 处理文件响应
const processFileResponseSchema = {
    type: "object",
    properties: {
        bcode: {
            type: "object",
            properties: {
                code: { type: "integer" },
                message: { type: "string" }
            },
            required: ["code"]
        },
        data: {
            type: "object",
            properties: {
                chunksProcessed: { type: "integer" }
            },
            required: ["chunksProcessed"]
        }
    },
    required: ["bcode", "data"]
};

// 获取文件列表响应
const getFilesResponseSchema = {
    type: "object",
    properties: {
        bcode: {
            type: "object",
            properties: {
                code: { type: "integer" },
                message: { type: "string" }
            },
            required: ["code"]
        },
        data: {
            type: "array",
            items: {
                type: "object",
                properties: {
                    id: { type: "string" },
                    sessionId: { type: "string" },
                    name: { type: "string" },
                    path: { type: "string" },
                    size: { type: "integer" },
                    type: { type: "string" },
                    createdAt: { type: "string" },
                    updatedAt: { type: "string" }
                },
                required: ["id", "sessionId", "name"]
            }
        }
    },
    required: ["bcode", "data"]
};

// 切换会话模型请求
const changeSessionModelRequestSchema = {
    type: "object",
    properties: {
        sessionId: { type: "string" },
        modelId: { type: "string" },
        embedModelId: { type: "string" }
    },
    required: ["sessionId", "modelId"]
};

// 切换会话模型响应
const changeSessionModelResponseSchema = {
    type: "object",
    properties: {
        bcode: {
            type: "object",
            properties: {
                code: { type: "integer" },
                message: { type: "string" }
            },
            required: ["code"]
        },
        data: {
            type: "object",
            properties: {
                id: { type: "string" },
                title: { type: "string" },
                modelId: { type: "string" },
                modelName: { type: "string" },
                embedModelId: { type: "string" },
                thinkingEnabled: { type: "boolean" },
                createdAt: { type: "string" },
                updatedAt: { type: "string" }
            },
            required: ["id", "modelId", "modelName"]
        }
    },
    required: ["bcode", "data"]
};

module.exports = {
    baseResponseSchema,
    createSessionRequestSchema,
    createSessionResponseSchema,
    getSessionsResponseSchema,
    sendMessageRequestSchema,
    sendMessageResponseSchema,
    sendStreamMessageRequestSchema,
    getMessagesRequestSchema,
    getMessagesResponseSchema,
    uploadFileResponseSchema,
    processFileRequestSchema,
    processFileResponseSchema,
    getFilesResponseSchema,
    changeSessionModelRequestSchema,
    changeSessionModelResponseSchema
};
