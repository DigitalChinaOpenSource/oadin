import { useState, useEffect } from 'react';
export default function useViewModel() {
  const [isUploadVisible, setIsUploadVisible] = useState(false);
  // TODO 获取当前是否下载词嵌入模型

  return {
    isUploadVisible,
    setIsUploadVisible,
  };
}
