import { useEffect, useState } from 'react';
import { IModelDataItem, IModelSourceType, ModelData } from '@/types';
import { httpRequest } from '@/utils/httpRequest';
import { useRequest } from 'ahooks';
import useModelListStore from '@/store/useModelListStore';
import { dealSmartVisionModels } from '@/components/model-manage-tab/model-list-content/utils.ts';

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
}

export function useViewModel(): IMyModelListViewModel {
  // 模型/问学列表全量数据
  const [modelListData, setModelListData] = useState<IModelDataItem[]>([]);

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
      if (paramsTemp.service_source === 'remote') {
        // 处理问学模型列表的数据, 把推荐的模型放在前面
        return dealSmartVisionModels(data?.data || []);
      }
      return data?.data || [];
    },
    {
      manual: true,
      onSuccess: (data) => {
        // 处理一些数据格式
        const dataWithSource = (data || []).map(
          (item) =>
            ({
              ...item,
              currentDownload: 0,
            }) as any,
        );
        setModelListData(dataWithSource);
      },
      onError: (error) => {
        console.error('获取模型列表失败:', error);
      },
    },
  );

  return {
    modelSupportLoading,
    modelListData,
    fetchModelSupport,
  };
}
