import { baseHeaders } from '../../utils/index';
import { fetchEventSource } from '@microsoft/fetch-event-source';
import { processStreamData, createStateTracker } from './streamProcessor';

const NO_DATA_TIMEOUT = 5000; // 5秒无数据就断开
const TOTAL_TIMEOUT = 10000; // 总超时时间

/**
 * 暂停模型下载
 * @param data - 请求体参数
 */
function abortDownload(data: any) {
  return fetch(`/byze/cancelDownload`, {
    method: 'POST',
    headers: baseHeaders() as any,
    body: JSON.stringify(data),
  })
    .then((res) => {
      return res.json();
    })
    .catch((e) => {
      console.error(e);
      return { models: null, error: e.message };
    });
}

/**
 * 设置无数据定时器
 * @param onTimeout - 超时时的回调
 * @param retryFn - 重试函数
 * @param hasRetriedRef - 是否已重试的引用
 */
function setupNoDataTimer(onTimeout: (err: any) => void, retryFn: () => void, hasRetriedRef: any) {
  return setTimeout(() => {
    if (!hasRetriedRef.current) {
      console.warn('首次5秒内无数据返回，开始重试...');
      hasRetriedRef.current = true;
      retryFn();
    } else {
      console.error('重试后仍未收到数据，触发总超时');
      onTimeout(new Error('流式请求超时：10秒未收到数据'));
    }
  }, NO_DATA_TIMEOUT);
}

/**
 * 清除所有的定时器
 * @param timers - 定时器数组
 */
function clearTimers(...timers: any[]) {
  timers.forEach((timer) => timer && clearTimeout(timer));
}

/**
 * 开始下载（带单次重试 + 总超时）
 * @param data - 请求体参数
 * @param options - 包含各种回调的对象
 */
async function startDownLoad(data: any, options: any) {
  let noDataTimer: any = null;
  let totalTimeoutId: any = null;
  const hasRetriedRef: any = { current: false };
  // 创建状态追踪器
  const stateTracker = createStateTracker();

  const resetNoDataTimer = () => {
    clearTimers(noDataTimer);
    noDataTimer = setupNoDataTimer(options.onerror, startFetch, hasRetriedRef);
  };

  async function startFetch() {
    const abortController = new AbortController();
    const signal = abortController.signal;

    // 处理请求参数，确保与后端接口兼容
    const requestData = {
      ...data,
      provider_name: data.serviceName === 'text_to_image' ? 'baidu' : data.providerName
    };

    await fetchEventSource(`/byze/installModelStream`, {
      method: 'POST',
      headers: baseHeaders(),
      body: JSON.stringify(requestData),
      openWhenHidden: true,
      signal,
      onopen: async (response: Response) => {
        // 连接打开时开始监听无数据超时
        resetNoDataTimer();
        options.onopen?.();
        console.log('Event source opened');
      },
      onmessage: (event) => {
        if (event.data && event.data !== '[DONE]') {
          try {
            // 使用新的流处理逻辑
            const currentState = stateTracker.getState();
            const parsedData = JSON.parse(event.data);
            const processedData = processStreamData(event, currentState);
            
            if (processedData) {
              // 更新状态追踪器
              stateTracker.update(parsedData, processedData);
              // 调用回调
              options.onmessage?.(processedData);
            }
            
            // 收到消息则重置无数据计时器
            resetNoDataTimer();
          } catch (err) {
            console.error('解析事件流失败:', err);
            options.onerror?.(new Error('事件流格式错误'));
          }
        }
      },
      onclose: () => {
        clearTimers(noDataTimer, totalTimeoutId);
        console.log('Event source closed');
        options.onclose?.();
      },
      onerror: (error) => {
        clearTimers(noDataTimer, totalTimeoutId);
        console.error('EventSource 错误:', error);
        options.onerror?.(error);
      },
    });
  }

  // 设置一个总的10秒超时定时器
  totalTimeoutId = setTimeout(() => {
    if (!hasRetriedRef.current) {
      console.warn('10秒总超时前强制进行一次重试');
      hasRetriedRef.current = true;
      startFetch();
    } else {
      console.error('总超时：10秒内未收到任何数据');
      options.onerror?.(new Error('流式请求超时：10秒未收到数据'));
    }
  }, TOTAL_TIMEOUT);

  // 首次启动请求
  await startFetch();
}

export { abortDownload, startDownLoad };
