import { useState, useEffect } from 'react';
import { useRequest } from '../../utils/useRequest';
import { ModelDataItem, ModelData } from './types';

export function useViewModel() {
  // 模型来源
  const [modelSourceVal, setModelSourceVal] = useState<string>('local');
  // 模型搜索
  const [modelSearchVal, setModelSearchVal] = useState<string>('');
  // 模型列表数据
  const [modelListData, setModelListData] = useState<ModelDataItem[]>([]);
  // 加载状态
  const [loading, setLoading] = useState<boolean>(false);

  const { get } = useRequest();

  // 监听 modelSourceVal 变化，只有当值为 local 或 all 时才请求数据
  useEffect(() => {
    if (modelSourceVal === 'local' || modelSourceVal === 'all') {
      fetchModelListData(modelSourceVal);
    }
  }, [modelSourceVal]);

  const onModelSourceChange = (val: string) => {
    setModelSourceVal(val);
  };

  const onModelSearch = (val: string) => {
    setModelSearchVal(val);
  };

  /**
   * 获取模型列表数据
   * @param serviceSource 服务来源，枚举值：local 或 remote
   */
  const fetchModelListData = async (serviceSource: string = 'local') => {
    try {
      setLoading(true);
      const data = await get<ModelData>('/model/support', { service_source: serviceSource, flavor: 'ollama' });
      setModelListData(data?.chat || []);
    } catch (error) {
      console.error('获取模型列表失败:', error);
      setModelListData([]);
    } finally {
      setLoading(false);
    }
  };

  return {
    modelSourceVal,
    onModelSourceChange,
    modelSearchVal,
    onModelSearch,
    modelListData,
    loading,
    fetchModelListData,
  };
}
