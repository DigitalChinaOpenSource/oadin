import { useEffect } from 'react';
import { message } from 'antd';
// 监听浏览器刷新 并执行某些操作
export const usePageRefreshListener = (onRefresh: () => void) => {
  useEffect(() => {
    const handleBeforeUnload = () => {
      onRefresh();
    };

    window.addEventListener('beforeunload', handleBeforeUnload);

    return () => {
      window.removeEventListener('beforeunload', handleBeforeUnload);
    };
  }, []);
};

// 检查是否达到下载数量限制
export const checkIsMaxDownloadCount = ({ modelList, downList, modelType, id }: any) => {
  // 检查是否存在可选项
  if (modelList) {
    // 已下载的直接跳过
    const hasDownloadableModels = modelList.find((item: any) => item.canSelect && item.id === id);
    if (hasDownloadableModels) return false;
  }

  // 检查是否已存在相同模型
  const isModelNotInList = !downList.some((item: any) => item?.modelType === modelType && item?.id === id);
  console.info(
    downList.filter((item: any) => !(item.canSelect && item.currentDownload === 100)),
    '过滤的下载列表',
  );
  // 验证下载数量限制
  const hasReachedLimit = downList.filter((item: any) => !(item.canSelect || item.currentDownload === 100)).length > 2;

  // 触发限制条件
  if (isModelNotInList && hasReachedLimit) {
    message.warning('您已达到同时下载模型数的上限（3 个）。若您想下载新模型，请先完成或取消部分现有下载任务。');
    return true;
  }
  return false;
};

export function eventStreamToAsyncIterator(emitter: any, endEvent = 'end') {
  const asyncIterable = {
    [Symbol.asyncIterator]() {
      const queue = [] as any[];
      let resolve: any;
      let reject: any;
      let done = false;

      const onData = (data: any) => {
        if (resolve) {
          resolve({ value: data, done: false });
          resolve = null;
        } else {
          queue.push(data);
        }
      };

      const onEnd = () => {
        done = true;
        if (resolve) {
          resolve({ value: undefined, done: true });
        }
      };

      const onError = (err: any) => {
        if (reject) reject(err);
      };
      emitter.on('data', onData);
      emitter.once(endEvent, onEnd);
      emitter.once('error', onError);

      return {
        next() {
          if (queue.length > 0) {
            return Promise.resolve({ value: queue.shift(), done: false });
          }
          if (done) {
            return Promise.resolve({ value: undefined, done: true });
          }
          return new Promise((res, rej) => {
            resolve = res;
            reject = rej;
          });
        },
        return() {
          emitter.removeListener('data', onData);
          emitter.removeListener(endEvent, onEnd);
          emitter.removeListener('error', onError);
          return Promise.resolve({ done: true });
        },
      };
    },
  };
  return asyncIterable;
}
