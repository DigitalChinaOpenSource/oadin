import useOadinServerCheckStore from '@/store/useOadinServerCheckStore';

/**
 * 在每次请求前先检查健康状态
 * @param requestFn 需要执行的请求函数
 * @returns Promise<any>
 */
export async function requestWithHealthCheck<T>(requestFn: () => Promise<T>): Promise<T | undefined> {
  const { fetchOadinServerStatus, checkOadinStatus } = useOadinServerCheckStore.getState();

  // 先发起健康检查
  await fetchOadinServerStatus();

  // 检查健康状态
  if (!checkOadinStatus) {
    // 健康检查未通过，直接中断
    return Promise.reject(new Error('奥丁服务不可用，已拦截请求'));
  }

  // 健康检查通过，执行后续请求
  return requestFn();
}
