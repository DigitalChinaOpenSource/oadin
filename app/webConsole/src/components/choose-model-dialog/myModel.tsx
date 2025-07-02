import ModelManageTab from '@/components/model-manage-tab';
import React, { useEffect, useMemo, useRef } from 'react';
import { useViewModel as useViewModelContent } from '@/components/model-manage-tab/model-list-content/view-model.ts';
import { IUseSearchViewModelReturn, useViewModel } from '@/components/model-manage-tab/view-model.ts';
import { ISelectedDialogProps } from '@/components/choose-model-dialog/index.tsx';
import useModelListStore from '@/store/useModelListStore';
import useModelDownloadStore from '@/store/useModelDownloadStore';
import { IModelDataItem } from '@/types';

export const MyModel = (props: ISelectedDialogProps) => {
  const vmSearch: IUseSearchViewModelReturn = useViewModel();

  // 获取专门存储我的模型的状态和下载列表状态
  const { myModelsList, setMyModelsList } = useModelListStore();
  const { downloadList } = useModelDownloadStore();

  // 弹窗打开时，使用节流同步下载状态到我的模型列表
  const lastSyncTimeRef = useRef<number>(0);
  const pendingUpdateRef = useRef<boolean>(false);

  useEffect(() => {
    // 如果列表为空，不需要更新
    if (!myModelsList?.length || !downloadList?.length) return;

    // 使用节流机制，至少间隔500ms才更新一次
    const now = Date.now();
    if (now - lastSyncTimeRef.current < 500) {
      // 如果更新太频繁，我们只标记需要更新，不立即执行
      if (!pendingUpdateRef.current) {
        pendingUpdateRef.current = true;
        setTimeout(() => {
          syncDownloadStatus();
          pendingUpdateRef.current = false;
        }, 500);
      }
      return;
    }

    syncDownloadStatus();

    function syncDownloadStatus() {
      // 只获取需要更新的模型ID列表，避免不必要的全量更新
      const modelsNeedUpdate: string[] = [];

      // 找出需要更新的模型
      myModelsList.forEach((model) => {
        if (!model.id) return;

        const downloadItem = downloadList.find((item) => item.id === model.id);
        if (downloadItem && (model.status !== downloadItem.status || model.currentDownload !== downloadItem.currentDownload || model.can_select !== downloadItem.can_select)) {
          modelsNeedUpdate.push(model.id);
        }
      });

      // 只有当有模型需要更新时才更新状态
      if (modelsNeedUpdate.length > 0) {
        setMyModelsList((currentList) =>
          currentList.map((model) => {
            if (!model.id || !modelsNeedUpdate.includes(model.id)) return model;

            const downloadItem = downloadList.find((item) => item.id === model.id);
            if (downloadItem) {
              return {
                ...model,
                status: downloadItem.status,
                currentDownload: downloadItem.currentDownload,
                can_select: downloadItem.can_select,
              };
            }
            return model;
          }),
        );
      }

      lastSyncTimeRef.current = Date.now();
    }
  }, [myModelsList, downloadList, props.isDialog]);

  // 获取接口数据内容，使用自定义的setListData函数
  const vmContent = useViewModelContent({
    onModelSearch: vmSearch.onModelSearch,
    modelSearchVal: vmSearch.modelSearchVal,
    modelSourceVal: vmSearch.modelSourceVal,
    mine: true,
    pageType: props.isDialog ? 1 : 0,
    // 使用memo确保函数稳定性
    customSetListData: useMemo(
      () => (list) => {
        // 同时更新专用的myModelsList
        setMyModelsList(list);
      },
      [setMyModelsList],
    ),
    // 提供当前存储的模型列表数据
    customModelListData: myModelsList,
  });

  return (
    <ModelManageTab
      {...props}
      vmContent={vmContent}
      vmSearch={vmSearch}
      currentTab="my-models"
      isDialog={props.isDialog}
    />
  );
};
