import { useEffect, useState } from 'react';
import { IModelDataItem, ModelData } from '@/types';
import { httpRequest } from '@/utils/httpRequest';
import { useRequest } from 'ahooks';
import useModelListStore from '@/store/useModelListStore';

export type ModelSourceType = 'local' | 'remote';

interface IModelSquareParams {
  flavor?: string;
  // remote时需要传
  // 'dev' | 'product'
  env_type?: string;
  service_source: ModelSourceType;
}

export interface IMyModelListViewModel {
  modelSupportLoading: boolean;
  modelListData: IModelDataItem[];
  fetchModelSupport: (params: IModelSquareParams) => void;
  fetchChangeModel: (params: { sessionId: string; modelId: string }) => void;
  fetchChooseModelNotify: (params: { service_name: string; local_provider?: string; remote_provider?: string }) => void;
}

export function useViewModel(): IMyModelListViewModel {
  // 直接从全局状态获取模型列表数据和更新函数
  const { myModelsList, setMyModelsList } = useModelListStore();
  // 本地状态，用于保存过滤后的模型列表
  const [modelListData, setModelListData] = useState<IModelDataItem[]>([]);
  
  // 初始化：如果myModelsList为空，则主动请求一次数据
  useEffect(() => {
    if (myModelsList.length === 0) {
      console.log("初始化：myModelsList为空，主动请求模型列表数据");
      // 这个会在下面的fetchModelSupport被调用
    } else {
      console.log("初始化：myModelsList已有数据，长度:", myModelsList.length);
      // 直接使用已有数据
      setModelListData(myModelsList);
    }
  }, []);

  // 获取模型列表 （本地和云端）
  const { loading: modelSupportLoading, run: fetchModelSupport } = useRequest(
    async (params: IModelSquareParams) => {
      const paramsTemp = {
        ...params,
        page_size: 999,
        mine: true,
      };
      if (params?.service_source === 'remote') {
        paramsTemp.env_type = 'product';
      }
      const data = await httpRequest.get<ModelData>('/control_panel/model/square', paramsTemp);

      return data?.data || [];
    },
    {
      manual: true,
      onSuccess: (data) => {
        // 处理一些数据格式
        let dataWithSource = (data || [])
          .map(
            (item) =>
              ({
                ...item,
                currentDownload: 0,
              }) as any,
          )
          .filter((item) => {
            return item.class.every((c_item: string) => !c_item.includes('嵌入'));
          });
        
        // 更新本地状态和全局状态
        console.log("请求模型成功，更新本地和全局状态，数据长度:", dataWithSource.length);
        setModelListData(dataWithSource);
        setMyModelsList(dataWithSource); // 同时更新全局状态
      },
      onError: (error) => {
        console.error('获取模型列表失败:', error);
      },
    },
  );

  // 切换会话模型
  const { run: fetchChangeModel } = useRequest(async (params: { sessionId: string; modelId: string }) => {
    if (!params?.sessionId || !params.modelId) {
      return {};
    }
    const data = await httpRequest.post('/playground/session/model', {
      ...params,
    });
    return data?.data || {};
  });

  // 选择模型后需要将所选择的通知奥丁
  const { run: fetchChooseModelNotify } = useRequest(async (params: { service_name: string; local_provider?: string; remote_provider?: string }) => {
    if (!params?.service_name) return;
    const data = await httpRequest.put('/service', {
      ...params,
    });
    return data || {};
  });

  // 监听全局myModelsList变化，更新本地状态
  useEffect(() => {
    console.log("ModelChecking - myModelsList 变化，长度:", myModelsList.length);
    
    // 总是更新本地状态，即使列表为空
    // 这样可以确保UI总是反映最新的数据状态
    setModelListData(myModelsList);
    
    // 记录可用的模型数量（仅用于调试）
    const availableModels = myModelsList.filter(model => 
      model.status === 2 && model.can_select
    );
    console.log("ModelChecking - 可用完成状态模型:", availableModels.length);
  }, [myModelsList]);

  return {
    modelSupportLoading,
    modelListData,
    fetchModelSupport,
    fetchChooseModelNotify,
    fetchChangeModel,
  };
}
