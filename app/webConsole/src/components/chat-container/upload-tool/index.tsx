import React, { useState } from 'react';
import { Button, Upload, Tooltip, message } from 'antd';
import type { UploadProps, UploadFile } from 'antd';
import { httpRequest } from '@/utils/httpRequest';
import uploadSvg from '@/components/icons/upload.svg';
import useChatStore from '../store/useChatStore';
import { getSessionIdFromUrl, setSessionIdToUrl } from '@/utils/sessionParamUtils';

interface UploadToolProps {
  maxFiles?: number;
  maxFileSize?: number;
}

// 自定义文件状态类型
export type FileStatus = 'error' | 'uploading' | 'done';

export default function UploadTool({ maxFiles = 5, maxFileSize = 50 }: UploadToolProps) {
  const { uploadFileList, setUploadFileList } = useChatStore();
  // 从URL中获取当前会话ID
  const currentSessionId = getSessionIdFromUrl();
  const maxFileSizeBytes = maxFileSize * 1024 * 1024;
  const validateFile = (
    file: File,
  ): {
    isValid: boolean;
    errorType?: 'format' | 'size';
    errorMessage?: string;
  } => {
    const supportedFormats = ['txt', 'html', 'md', 'markdown', 'pdf', 'doc', 'docx', 'xls', 'xlsx'];
    const fileName = file.name.toLowerCase();
    const fileExtension = fileName.slice(fileName.lastIndexOf('.') + 1);
    const isFormatValid = supportedFormats.includes(fileExtension);
    if (!isFormatValid) {
      return {
        isValid: false,
        errorType: 'format',
        errorMessage: '不支持的文件格式',
      };
    }

    const isSizeValid = file.size <= maxFileSizeBytes;
    if (!isSizeValid) {
      return {
        isValid: false,
        errorType: 'size',
        errorMessage: `文件大小超过${maxFileSize}MB限制`,
      };
    }

    return {
      isValid: true,
    };
  };

  const handleBeforeUpload = (file: File) => {
    if (uploadFileList && uploadFileList.length >= maxFiles) {
      message.error(`最多只能上传${maxFiles}个文件`);
      return false;
    }
    const validationResult = validateFile(file);
    if (!validationResult.isValid) {
      message.error(validationResult.errorMessage);
      return false;
    }
    return true;
  };

  // 自定义上传请求
  const customUploadRequest: UploadProps['customRequest'] = async ({ file, onProgress, onSuccess, onError }) => {
    // 转换为File对象
    const fileObj = file as File;

    // 创建一个文件对象，状态为uploading
    const uploadingFile: UploadFile = {
      uid: `${Date.now()}-${Math.random()}`,
      name: fileObj.name,
      status: 'uploading', // 初始状态为上传中
      percent: 0,
      size: fileObj.size,
      type: fileObj.type,
    };

    // 立即更新父组件状态，显示正在上传
    const initialFileList = [...uploadFileList, uploadingFile];
    setUploadFileList(initialFileList);

    const formData = new FormData();
    formData.append('file', fileObj);
    formData.append('sessionId', currentSessionId || '');
    try {
      // 发送请求
      const response = await httpRequest.post('/playground/file', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });
      const responseData = response.data?.data || response.data || response;
      const successFile: UploadFile = {
        ...uploadingFile,
        status: 'done',
        percent: 100,
        response: responseData, // 只保存处理后的数据
        url: responseData.url || '', // 如果响应中包含文件URL
      };

      // 获取当前最新的文件列表状态
      const currentFileList = useChatStore.getState().uploadFileList;

      // 更新列表中的文件状态
      const newFileList = currentFileList
        .map((item) => (item.uid === uploadingFile.uid ? successFile : item))
        .filter(
          (item, index, self) =>
            // 过滤掉重复项
            self.findIndex((f) => f.uid === item.uid) === index,
        );

      setUploadFileList(newFileList);
      onSuccess?.(responseData);

      message.success(`文件 ${fileObj.name} 上传成功`);
    } catch (error: Error | any) {
      const currentFileList = useChatStore.getState().uploadFileList;
      const failedFile: UploadFile = {
        ...uploadingFile,
        status: 'error',
        error: error as Error,
      };
      const newFileList = currentFileList.map((item) => (item.uid === uploadingFile.uid ? failedFile : item));
      setUploadFileList(newFileList);
      onError?.(error as Error);
      message.error(`文件 ${fileObj.name} 上传失败: ${error?.message || '未知错误'}`);
    }
  };

  return (
    <Upload
      showUploadList={false} // 不显示默认上传列表，由父组件自定义展示
      fileList={uploadFileList}
      beforeUpload={handleBeforeUpload}
      customRequest={customUploadRequest}
      multiple={false}
      accept=".txt,.html,.htm,.md,.markdown,.pdf,.doc,.docx,.xls,.xlsx"
    >
      <Tooltip title="文件格式支持 txt、HTML、Markdown、PDF、DOC、DOCX、XLS、XLSX，单个文件限制 50MB">
        <Button
          icon={
            <img
              src={uploadSvg}
              alt="上传"
            />
          }
          disabled={uploadFileList.some((file) => file.status === 'uploading')} // 如果有文件正在上传，禁用按钮
        />
      </Tooltip>
    </Upload>
  );
}
