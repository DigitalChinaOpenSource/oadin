import { useState, useCallback } from 'react';
import { IModelAuthorize, IModelAuthType } from '../types';
import { Modal } from 'antd';

const { confirm } = Modal;

export function useViewModel() {
  // 模型存储路径弹窗是否显示
  const [modalPathVisible, setModalPathVisible] = useState<boolean>(false);
  // 接口获取
  const [modelPath, setModelPath] = useState<string>('');
  // 云端模型授权弹窗是否显示
  const [modelAuthVisible, setModelAuthVisible] = useState<boolean>(false);
  // 云端模型授权
  const [modelAuthorize, setModelAuthorize] = useState<IModelAuthorize>({
    apiHost: '',
    apiKey: '',
  });
  // 模型详情弹窗是否显示
  const [isDetailVisible, setIsDetailVisible] = useState(false);
  // 配置 ｜ 更新授权
  const [modelAuthType, setModelAuthType] = useState<IModelAuthType>('config');

  // 模型详情弹窗
  const onDetailModalVisible = useCallback((visible: boolean) => {
    setIsDetailVisible(visible);
  }, []);

  //
  const onModelPathVisible = useCallback(() => {
    setModalPathVisible(!modalPathVisible);
  }, [modalPathVisible]);

  const fetchModalPath = () => {
    setModelPath('URL_ADDRESS.baidu.com');
  };

  const onModelAuthVisible = useCallback((visible: boolean, type: IModelAuthType) => {
    setModelAuthVisible(visible);
    setModelAuthType(type);
  }, []);

  const onSetModelAuthorize = (authData: IModelAuthorize) => {
    setModelAuthorize(authData);
  };

  const downloadConfirm = (modelData: any) => {
    confirm({
      title: '确认下载此模型？',
      okText: '确认下载',
      onOk() {
        console.log('OK');
      },
      onCancel() {
        console.log('Cancel');
      },
    });
  };

  const deleteConfirm = (modelData: any) => {
    confirm({
      title: '是否确认删除此模型',
      content: '此操作将删除您与该模型的所有对话记录、文件信息，并终止相关进程。',
      okButtonProps: {
        style: { backgroundColor: '#e85951' },
      },
      okText: '确认删除',
      onOk() {
        console.log('OK');
      },
      onCancel() {
        console.log('Cancel');
      },
    });
  };

  return {
    modelPath,
    modalPathVisible,
    onModelPathVisible,

    fetchModalPath,
    modelAuthorize,
    onSetModelAuthorize,

    modelAuthVisible,
    onModelAuthVisible,

    isDetailVisible,
    onDetailModalVisible,
    modelAuthType,

    deleteConfirm,
    downloadConfirm,
  };
}
