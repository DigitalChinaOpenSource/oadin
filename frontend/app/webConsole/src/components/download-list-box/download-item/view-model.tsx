import useModelDownloadStore from '@/store/useModelDownloadStore';
import { LOCAL_STORAGE_KEYS } from '@/constants';
import { useDownLoad } from '@/hooks/useDownload';
import { message } from 'antd';
import { IModelDataItem } from '@/types';
export function useViewModel() {
  const { fetchDownLoadAbort, fetchDownloadStart } = useDownLoad();
  const { setDownloadList, downloadList } = useModelDownloadStore();
  const fetchCancelModel = async (data: IModelDataItem) => {
    await fetchDownLoadAbort({ model_name: data.name }, { id: data.id });
  };

  const fetchRemoveModel = async (data: IModelDataItem) => {
    if (!downloadList.some((item) => item.name === data.name)) {
      message.warning('未找到匹配的模型进行移除');
      return;
    }
    const result = downloadList.filter((item) => item.name !== data.name);
    setDownloadList(result);
    localStorage.setItem(LOCAL_STORAGE_KEYS.MODEL_DOWNLOAD_LIST, JSON.stringify(result));

    await fetchCancelModel(data);
  };

  return { fetchCancelModel, fetchDownloadStart, fetchRemoveModel };
}
