import { useState } from "react";
export function useViewModel() {
   const [baseInfo, setBaseInfo] = useState({
        name: '白泽服务提供商',
        status: '启用',
        statusCode: '1'
    })
    const [modelList, setModelList] = useState([{name: '模型1', tags: ['深度思考', '文本模型'], size: 32, status: 1, text: '深度求索'}])

    return { baseInfo, modelList }
}