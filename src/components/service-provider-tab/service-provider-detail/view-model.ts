import { useState } from "react";
export function useViewModel() {
   const [baseInfo, setBaseInfo] = useState({
        name: '白泽服务提供商',
        status: '启用',
        createTime: '2023-10-01 12:00:00',
        updateTime: '2023-10-01 12:00:00',
        statusCode: '1'
    })
    const [modelList, setModelList] = useState([{name: '模型1', tags: ['深度思考', '文本模型'], size: 32, status: 1, text: '深度求索'}, {name: '模型1', tags: ['深度思考', '文本模型'], size: 32, status: 1, text: '深度求索'}])
    const [pagination, setPagination] = useState({current: 1, pageSize: 10, total: 20})
    const handlePagChange = (page: number) => {
        console.log(page)
        setPagination({ ...pagination, current: page })
    }
    return { baseInfo, modelList, pagination, handlePagChange }
}