package byze

import "github.com/invopop/jsonschema"

func GenerateSchema(obj interface{}) *jsonschema.Schema {
    reflector := &jsonschema.Reflector{
        // 自定义配置（可选）
        AllowAdditionalProperties:  true,
        RequiredFromJSONSchemaTags: true,
    }
    return reflector.Reflect(obj)
}

// 生成UserRequest的Schema
userReqSchema := GenerateSchema(UserRequest{})