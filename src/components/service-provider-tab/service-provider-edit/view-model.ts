import { useState } from "react";
export function useViewModel() {
    const [modelList, setModelList] = useState([{name: '模型1', key: '1'}, {name: '模型1', key: '2'}])
    const [serviceOrgList, setServiceOrgList] = useState([{name: '本地', key: '1'}, {name: '服务提供商2', key: '2'} ])
   
    return { modelList, serviceOrgList }
}