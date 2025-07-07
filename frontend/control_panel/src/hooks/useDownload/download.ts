//*****************************************************************************
// Copyright 2025 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//*****************************************************************************

import { baseHeaders } from '../../utils/index';
import { fetchEventSource } from '@microsoft/fetch-event-source';
import { httpRequest } from '../../utils/httpRequest';
import { IRequestModelParams } from '../../types';
import { IDownParseData, IDownloadCallbacks, IProgressData } from './types';
import { API_PREFIX } from '@/constants';

/**
 * 暂停模型下载
 * @param data - 请求体参数
 */
async function abortDownload(data: { model_name: string }) {
  // console.log('abortDownload', data);
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
async function modelDownloadStream(data: IRequestModelParams, { onmessage, onerror, onopen, onclose }: IDownloadCallbacks) {
  let noDataTimer: NodeJS.Timeout | null = null;
  let totalTimeoutId: NodeJS.Timeout | null = null;
  let hasRetried = false; // 是否已重试一次
  const NO_DATA_TIMEOUT = 10000;
  const TOTAL_TIMEOUT = 30000;

  // 状态变量
  let lastUsefulDataObj: IDownParseData | null = null;
  let lastCompleted = 0;
  let overallProgress = 0;
  let lastDigest: string | null = null;
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

  const sendCompletionMessage = (lastUsefulDataObj: IDownParseData) => {
    const finalDataObj = {
      progress: 100,
      status: 'success',
      completedsize: lastUsefulDataObj ? lastUsefulDataObj.completedsize : 0,
      totalsize: lastUsefulDataObj ? lastUsefulDataObj.totalsize : 0,
    };
    onmessage?.(finalDataObj);
    console.log('模型下载完成:', finalDataObj);
  };

  const processProgressData = (part: IProgressData) => {
    let dataObj = { status: part.status } as IDownParseData;

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
        } as IDownParseData;

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
    const API_BASE_URL = import.meta.env.VITE_HEALTH_API_URL || '';

    fetchEventSource(`${API_BASE_URL}${API_PREFIX}/model/stream`, {
      method: 'POST',
      headers: baseHeaders(),
      body: JSON.stringify(data),
      openWhenHidden: true,
      signal,
      onmessage: (event) => {
        if (event.data && event.data !== '[DONE]') {
          try {
            const parsedData = JSON.parse(event.data) as IDownParseData;

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
                message: parsedData?.message || parsedData?.data || 'Download error',
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
              sendCompletionMessage(lastUsefulDataObj as IDownParseData);
              return;
            }
            // 处理进度数据
            const dataObj = processProgressData(parsedData) as IDownParseData;
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
    console.error('总超时：30秒内未收到任何数据');
    onerror?.(new Error('流式请求超时：30秒未收到数据'));
  }, TOTAL_TIMEOUT);

  startFetch();
}

export { abortDownload, modelDownloadStream };
