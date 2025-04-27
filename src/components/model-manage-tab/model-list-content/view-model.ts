import { useState } from "react";

export function useViewModel() {
  // 模型存储路径弹窗是否显示
  const [modalPathVisible, setModalPathVisible] = useState<boolean>(false);
  // 接口获取
  const [modelPath, setModelPath] = useState<string>('');

  const onModelPathVisible = () => {
    setModalPathVisible(!modalPathVisible);
  }

  const fetchModalPath = () => {
    setModelPath('URL_ADDRESS.baidu.com');
  }

	return {
    modelPath,
    modalPathVisible,
    onModelPathVisible,
    fetchModalPath
	}
}