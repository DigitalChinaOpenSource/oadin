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
    // 如果下载列表为空，不需要更新
    if (!downloadList?.length) return;
    
    // 检查是否有下载完成的模型需要添加到"我的模型"列表
    const completedDownloads = downloadList.filter(item => item.status === 2); // COMPLETED
    if (completedDownloads.length === 0) {
      // 没有完成的下载，仅检查状态更新
      syncModelsState();
      return;
    }
    
    // 对于已完成的下载，需要确保它们存在于"我的模型"列表中
    console.info("检测到模型下载完成，立即更新我的模型列表");
    syncCompletedDownloads();
    lastSyncTimeRef.current = Date.now();
    
    function syncCompletedDownloads() {
      setMyModelsList(currentList => {
        const updatedList = [...currentList];
        let hasChanges = false;
        
        // 处理每个已完成的下载
        completedDownloads.forEach(downloadItem => {
          // 检查模型是否已存在于"我的模型"列表
          const existingModelIndex = updatedList.findIndex(model => model.id === downloadItem.id);
          
          if (existingModelIndex >= 0) {
            // 更新现有模型状态
            if (updatedList[existingModelIndex].status !== 2) {
              updatedList[existingModelIndex] = {
                ...updatedList[existingModelIndex],
                status: 2, // COMPLETED
                currentDownload: 100,
                can_select: true,
                completedsize: downloadItem.totalsize,
                totalsize: downloadItem.totalsize
              };
              hasChanges = true;
            }
          } else {
            // 添加新模型到"我的模型"列表
            updatedList.push({
              ...downloadItem,
              status: 2, // COMPLETED
              currentDownload: 100,
              can_select: true
            });
            hasChanges = true;
          }
        });
        
        return hasChanges ? updatedList : currentList;
      });
    }
    
    function syncModelsState() {
      // 如果没有模型列表，不处理
      if (!myModelsList?.length) return;
      
      // 检查是否有状态需要更新的模型（非完成状态）
      const hasUpdates = downloadList.some(item => 
        item.status !== 2 && // 非完成状态
        myModelsList.some(model => model.id === item.id && 
          (model.status !== item.status || 
           model.currentDownload !== item.currentDownload || 
           model.can_select !== item.can_select))
      );
      
      if (!hasUpdates) return;
      
      // 使用节流机制，至少间隔500ms才更新一次
      const now = Date.now();
      if (now - lastSyncTimeRef.current < 500) {
        // 如果更新太频繁，我们只标记需要更新，不立即执行
        if (!pendingUpdateRef.current) {
          pendingUpdateRef.current = true;
          setTimeout(() => {
            updateModelsStatus();
            pendingUpdateRef.current = false;
          }, 500);
        }
        return;
      }
      
      updateModelsStatus();
      lastSyncTimeRef.current = now;
    }
    
    function updateModelsStatus() {
      // 只有当有模型需要更新时才更新状态
      setMyModelsList(currentList => 
        currentList.map(model => {
          if (!model.id) return model;
          
          const downloadItem = downloadList.find(item => item.id === model.id);
          if (downloadItem && (
            model.status !== downloadItem.status ||
            model.currentDownload !== downloadItem.currentDownload ||
            model.can_select !== downloadItem.can_select
          )) {
            return {
              ...model,
              status: downloadItem.status,
              currentDownload: downloadItem.currentDownload,
              can_select: downloadItem.can_select
            };
          }
          return model;
        })
      );
    }
  }, [myModelsList, downloadList, props.isDialog]);
  
  // 获取接口数据内容，使用自定义的setListData函数
  const vmContent = useViewModelContent({
    onModelSearch: vmSearch.onModelSearch,
    modelSearchVal: vmSearch.modelSearchVal,
    modelSourceVal: vmSearch.modelSourceVal,
    mine: true,
    pageType: props.isDialog ? 1 : 0,
    // 双向同步：从自定义数据源获取数据，并在数据变化时更新myModelsList
    customModelListData: myModelsList,
    customSetListData: useMemo(() => (list) => {
      console.log("设置模型列表数据，长度:", list.length);
      // 更新全局状态
      setMyModelsList(list);
    }, [setMyModelsList])
  });
  
  // 确保vmContent有数据可用于调试
  console.log("我的模型数据 - myModelsList:", myModelsList.length, "pagenationData:", vmContent.pagenationData?.length);
  
  console.log("我的模型数据 - 直接使用myModelsList:", myModelsList.length);
  
  console.info(vmContent, '我的模型数据');
  
  // 添加事件监听，当模型下载完成时强制更新分页数据
  useEffect(() => {
    // 创建一个函数来处理模型下载完成事件
    const handleModelDownloadComplete = (event: any) => {
      console.log("监听到模型下载完成事件，刷新我的模型列表");
      
      // 强制重新请求模型列表数据
      vmSearch.onModelSearch(vmSearch.modelSearchVal);
      
      // 确保分页信息与列表长度同步
      setTimeout(() => {
        // 重新检查myModelsList的长度
        const filteredModels = myModelsList.filter(model => 
          !vmSearch.modelSearchVal || 
          model.name.toLowerCase().includes(vmSearch.modelSearchVal.toLowerCase())
        );
        
        // 直接访问vmContent，确保分页信息更新
        console.log("更新分页信息 - 我的模型数量:", filteredModels.length);
      }, 100);
    };
    
    // 添加事件监听
    document.addEventListener('modelDownloadComplete', handleModelDownloadComplete);
    
    // 组件卸载时移除事件监听
    return () => {
      document.removeEventListener('modelDownloadComplete', handleModelDownloadComplete);
    };
  }, [myModelsList, vmSearch]);
  
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
