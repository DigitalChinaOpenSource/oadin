import { useRequest } from 'ahooks';
import useModelDownloadStore from '@/store/useModelDownloadStore';
import { LOCAL_STORAGE_KEYS } from '@/constants';
import { useDownLoad } from '@/hooks/useDownload';
import { ModelDataItem } from '@/types';
export function useViewModel() {
  const { downLoadAbort, downLoadStart } = useDownLoad();
  const { downloadList, setDownloadList } = useModelDownloadStore();
  const { run: fetchCancelModelStream } = useRequest(downLoadAbort, {
    manual: true,
  });

  const fetchCancelModel = async (data: ModelDataItem) => {
    await fetchCancelModelStream(data.name, { id: data.id, modelType: data.modelType });
  };

  const fetchDownloadModel = async (data: ModelDataItem) => {
    await downLoadStart(data);
  };

  const fetchRemoveModel = async (data: ModelDataItem) => {
    const result = downloadList.filter((item) => item.name !== data.name);
    if (result.length === 0) return;
    setDownloadList(result);
    localStorage.setItem(LOCAL_STORAGE_KEYS.DOWN_LIST, JSON.stringify(result));

    await fetchCancelModel(data);
  };
  return { fetchCancelModel, fetchDownloadModel, fetchRemoveModel };
}
