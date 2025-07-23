import { useState, useEffect } from 'react';
import { useRequest } from 'ahooks';
import { httpRequest } from '@/utils/httpRequest';
import { IProviderDetailData, IProviderDetailParams } from '../types';
import { IServiceProviderDetailProps } from './index';
import dayjs from 'dayjs';
export function useViewModel(props: IServiceProviderDetailProps) {
  const { selectedRow, onCancel } = props;
  const [providerDetail, setProviderDetail] = useState<IProviderDetailData>({} as IProviderDetailData);

  useEffect(() => {
    fetchProviderDetail({ provider_name: selectedRow.provider_name });
  }, [selectedRow]);

  const { loading: providerDetailLoading, run: fetchProviderDetail } = useRequest(
    async (params: IProviderDetailParams) => {
      const paramsTemp = {
        provider_name: selectedRow.provider_name,
        page: params.page || 1,
        page_size: params.page_size || 5,
      } as IProviderDetailParams;
      if (params.provider_name.includes('smartvision')) {
        paramsTemp.env_type = 'product';
      }
      const data = await httpRequest.get<IProviderDetailData>('/service_provider/detail', { ...paramsTemp });
      return data || {};
    },
    {
      manual: true,
      onSuccess: (data) => {
        if (!data) {
          setProviderDetail(selectedRow as any);
          return;
        }
        setProviderDetail(data);
      },
      onError: (error) => {
        setProviderDetail(selectedRow as any);
        console.error('获取服务提供商数据失败:', error);
      },
    },
  );

  const handlePageChange = (page: number, page_size: number) => {
    fetchProviderDetail({
      provider_name: selectedRow.provider_name,
      page,
      page_size: 5,
    });
  };

  const formateIsoTime = (isoTime: string) => {
    if (!isoTime) return null;
    return dayjs(isoTime).format('YYYY-MM-DD HH:mm:ss');
  };

  return { providerDetailLoading, providerDetail, selectedRow, onCancel, formateIsoTime, handlePageChange };
}
