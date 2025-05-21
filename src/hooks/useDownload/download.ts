import { baseHeaders } from '../../utils/index';
import { fetchEventSource } from '@microsoft/fetch-event-source';
import { httpRequest } from '../../utils/httpRequest';
import { IRequestModelParams } from '../../types';

/**
 * 暂停模型下载
 * @param data - 请求体参数
 */
async function abortDownload(data: { model_name: string }) {
  console.log('abortDownload', data);
  await httpRequest
    .post('/model/stream/cancel', data)
    .then((res) => res)
    .catch((e) => {
      console.error(e);
      return { models: null, error: e.message };
    });
}

/**
 * 启动模型下载任务，处理流式数据并返回进度和状态
 */
async function modelDownloadStream(data: IRequestModelParams, { onmessage, onerror, onopen, onclose }: any) {
  let noDataTimer: any = null;
  let totalTimeoutId: any = null;
  let hasRetried = false; // 是否已重试一次
  const NO_DATA_TIMEOUT = 10000;
  const TOTAL_TIMEOUT = 20000;

  // 状态变量
  let lastUsefulDataObj: any = null;
  let lastCompleted = 0;
  let overallProgress = 0;
  let lastDigest: any = null;
  let isFirstChunk = true;

  const clearTimers = () => {
    if (noDataTimer) clearTimeout(noDataTimer);
    if (totalTimeoutId) clearTimeout(totalTimeoutId);
  };

  const resetNoDataTimer = () => {
    clearTimers();
    noDataTimer = setTimeout(() => {
      if (!hasRetried) {
        hasRetried = true;
        clearTimers();
        startFetch();
      } else {
        console.error('重试后仍未收到数据，触发总超时');
        onerror?.(new Error('流式请求超时：10秒未收到数据'));
      }
    }, NO_DATA_TIMEOUT);
  };

  const sendCompletionMessage = (lastUsefulDataObj: any) => {
    const finalDataObj = {
      progress: 100,
      status: 'success',
      completedsize: lastUsefulDataObj ? lastUsefulDataObj.completedsize : 0,
      totalsize: lastUsefulDataObj ? lastUsefulDataObj.totalsize : 0,
    };
    onmessage?.(finalDataObj);
    console.log('模型下载完成:', finalDataObj);
  };

  const processProgressData = (part: any) => {
    let dataObj = { status: part.status };

    if (part?.digest) {
      let percent = 0;
      if (part.completed && part.total) {
        const completed = Math.max(part.completed, lastCompleted);

        if (isFirstChunk) {
          // 第一个 chunk 占据总量的 94%
          percent = Math.round((completed / part.total) * 94);
        } else {
          // 其他 chunk 占据剩余的 6%
          // 检查是否是新的 digest
          if (part.digest !== lastDigest) {
            // 新的 chunk 开始，记录当前进度作为基础值
            isFirstChunk = false;
            // 不立即增加进度，等待该 chunk 的进度数据
          }

          // 计算当前 chunk 内的进度百分比（占总进度的 6%）
          const chunkPercentage = (completed / part.total) * 6;
          // 确保不超过总的 6%
          const boundedChunkPercentage = Math.min(chunkPercentage, 6);
          // 基础进度(94%) + 当前 chunk 进度(最多 6%)
          percent = 94 + boundedChunkPercentage;
          // 确保不超过 100%
          percent = Math.min(Math.round(percent), 100);
        }

        dataObj = {
          progress: percent,
          status: part.status,
          completedsize: Math.floor(completed / 1000000),
          totalsize: Math.floor(part.total / 1000000),
        } as any;

        // 更新状态
        lastCompleted = completed;
        overallProgress = percent;
        lastDigest = part.digest;
        lastUsefulDataObj = dataObj;
      }
    }

    return dataObj;
  };

  const startFetch = () => {
    const abortController = new AbortController();
    const signal = abortController.signal;

    fetchEventSource(`/byze/v0.2/model/stream`, {
      method: 'POST',
      headers: baseHeaders(),
      body: JSON.stringify(data),
      openWhenHidden: true,
      signal,
      onmessage: (event) => {
        if (event.data && event.data !== '[DONE]') {
          try {
            const parsedData = JSON.parse(event.data);
            console.log('接收到的事件流数据:', parsedData);
            if (totalTimeoutId) {
              clearTimeout(totalTimeoutId);
              totalTimeoutId = null;
            }
            if (noDataTimer) clearTimeout(noDataTimer);
            noDataTimer = setTimeout(() => {
              resetNoDataTimer();
            }, NO_DATA_TIMEOUT);
            // 处理错误
            if (parsedData?.status === 'error') {
              onmessage?.({
                status: 'error',
                message: parsedData.message || '模型下载失败',
              });
              return;
            }
            // 处理取消
            if (parsedData?.status === 'canceled') {
              onmessage?.({
                ...(lastUsefulDataObj || {}),
                status: 'canceled',
              });
              return;
            }
            // 处理成功
            if (parsedData?.status === 'success') {
              sendCompletionMessage(lastUsefulDataObj);
              return;
            }
            // 处理进度数据
            const dataObj = processProgressData(parsedData);
            console.log('模型下载进度:', dataObj);
            onmessage?.(dataObj);
            resetNoDataTimer();
          } catch (err) {
            console.error('解析事件流失败:', err);
            onerror?.(new Error('事件流格式错误'));
          }
        }
      },
      onerror: (error) => {
        clearTimers();
        console.error('EventSource 错误:', error);
        onerror?.(error);
      },
      // @ts-ignore
      onopen: () => {
        onopen?.();
        console.log('Event source opened');
      },
      onclose: () => {
        clearTimers();
        console.log('Event source closed');
        onclose?.();
      },
    });
  };

  totalTimeoutId = setTimeout(() => {
    console.error('总超时：20秒内未收到任何数据');
    onerror?.(new Error('流式请求超时：20秒未收到数据'));
  }, TOTAL_TIMEOUT);

  startFetch();
}

export { abortDownload, modelDownloadStream };
