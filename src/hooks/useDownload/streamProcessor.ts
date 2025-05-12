/**
 * 处理模型流数据
 * @param event - SSE 事件数据
 * @param lastState - 上一次的状态
 * @returns 处理后的数据对象
 */
export function processStreamData(event: any, lastState: any = {}) {
  if (!event || !event.data) {
    return null;
  }

  try {
    const part = JSON.parse(event.data);
    // 检查错误状态
    if (part?.status === 'error') {
      return {
        status: 'error',
        message: part?.message || '模型下载失败',
      };
    }
    // 检查取消状态
    if (part?.status === 'canceled') {
      return {
        ...(lastState.lastUsefulDataObj || {}),
        status: 'canceled',
      };
    }
    // 检查成功状态
    if (part?.status === 'success') {
      return {
        progress: 100,
        status: 'success',
        completedsize: lastState.lastUsefulDataObj ? lastState.lastUsefulDataObj.completedsize : 0,
        totalsize: lastState.lastUsefulDataObj ? lastState.lastUsefulDataObj.totalsize : 0,
      };
    }

    // 处理进度信息
    return processProgressData(part, lastState.lastDigest, lastState.lastCompleted, lastState.isFirstChunk, lastState.overallProgress);
  } catch (err) {
    console.error('解析事件流失败:', err);
    return {
      status: 'error',
      message: '事件流格式错误',
    };
  }
}

/**
 * 处理进度数据
 * @param part - 当前数据块
 * @param lastDigest - 上一个数据块的摘要
 * @param lastCompleted - 上一次完成的大小
 * @param isFirstChunk - 是否是第一个数据块
 * @param overallProgress - 总体进度
 * @returns 处理后的数据对象
 */
function processProgressData(part: any, lastDigest: any, lastCompleted = 0, isFirstChunk = true, overallProgress = 0) {
  let dataObj = { status: part.status } as any;

  if (part?.digest) {
    let percent = 0;
    if (part.completed && part.total) {
      const completed = Math.max(part.completed, lastCompleted);

      if (isFirstChunk) {
        // 第一个chunk占据总量的94%
        percent = Math.round((completed / part.total) * 94);
      } else {
        // 其他chunk占据总量的6%
        console.log(part.completed, part.total, '其他chunk 占据总量的6%');
        const chunkProgress = 6 / 6; // 每个chunk平均分配剩余的6%
        percent = overallProgress + chunkProgress;
      }

      dataObj = {
        progress: percent,
        status: part.status,
        completedsize: Math.floor(completed / 1000000),
        totalsize: Math.floor(part.total / 1000000),
      };
    }
  }

  return dataObj;
}

/**
 * 创建状态追踪器
 * @returns 状态追踪对象
 */
export function createStateTracker() {
  return {
    lastUsefulDataObj: null,
    lastCompleted: 0,
    overallProgress: 0,
    lastDigest: null,
    isFirstChunk: true,

    // 更新状态
    update(part: any, dataObj: any) {
      if (part?.digest) {
        if (this.lastDigest && part.digest !== this.lastDigest) {
          this.isFirstChunk = false;
        }
        this.lastDigest = part.digest;

        if (part.completed && part.total) {
          this.lastCompleted = Math.max(part.completed, this.lastCompleted);

          if (this.isFirstChunk) {
            this.overallProgress = Math.round((part.completed / part.total) * 94);
          } else {
            this.overallProgress += 6 / 6; // 其他chunk占据总量的6%，分6个等份
          }

          if (dataObj.completedsize) {
            this.lastUsefulDataObj = dataObj;
          }
        }
      }

      return dataObj;
    },

    // 获取当前状态
    getState() {
      return {
        lastUsefulDataObj: this.lastUsefulDataObj,
        lastCompleted: this.lastCompleted,
        overallProgress: this.overallProgress,
        lastDigest: this.lastDigest,
        isFirstChunk: this.isFirstChunk,
      };
    },
  };
}
