import React from 'react';
import { Button, Upload, Tooltip, message } from 'antd';
import type { UploadProps, UploadFile } from 'antd';
import uploadSvg from '@/components/icons/upload.svg';

interface UploadToolProps {
  uploadFileList: UploadFile[]; // 文件列表，由父组件维护
  onFileListChange: (fileList: UploadFile[]) => void; // 文件列表变更回调
  maxFiles?: number; // 最大上传文件数
}

export default function UploadTool({ onFileListChange, maxFiles = 5, uploadFileList }: UploadToolProps) {
  /**
   * 检查文件是否满足格式和大小限制
   * @param file 待检查的文件
   * @returns 包含验证结果和错误信息的对象
   */
  const validateFile = (
    file: File,
  ): {
    isValid: boolean;
    errorType?: 'format' | 'size';
    errorMessage?: string;
  } => {
    const supportedFormats = ['txt', 'html', 'htm', 'md', 'markdown', 'pdf', 'doc', 'docx', 'xls', 'xlsx', 'mp4', 'mp3', 'avi', 'wmv'];
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

    const maxSize = 50 * 1024 * 1024;
    // 检查文件大小
    const isSizeValid = file.size <= maxSize;
    if (!isSizeValid) {
      return {
        isValid: false,
        errorType: 'size',
        errorMessage: '文件大小超过50MB限制',
      };
    }

    return {
      isValid: true,
    };
  };

  const handleBeforeUpload = (file: File) => {
    // 检查文件数量限制
    if (uploadFileList && uploadFileList.length >= maxFiles) {
      message.error(`最多只能上传${maxFiles}个文件`);
      return false;
    }

    // 验证文件
    const validationResult = validateFile(file);
    if (!validationResult.isValid) {
      message.error(validationResult.errorMessage);
      return false;
    }
    // 通过验证的文件添加至fileList
    const newFile: UploadFile = {
      uid: `${Date.now()}-${Math.random()}`,
      name: file.name,
      status: 'done',
      size: file.size,
      type: file.type,
      // originFileObj: file as any,
    };
    onFileListChange([...uploadFileList, newFile]);
    return false;
  };

  const handleRemove = (file: UploadFile) => {
    // 直接调用父组件传递的回调，传入过滤后的文件列表
    onFileListChange(uploadFileList.filter((f) => f.uid !== file.uid));
    message.success(`已删除文件: ${file.name}`);
  };

  return (
    <Upload
      showUploadList={false}
      fileList={uploadFileList}
      beforeUpload={handleBeforeUpload}
      onRemove={handleRemove}
      accept=".txt,.html,.htm,.md,.markdown,.pdf,.doc,.docx,.xls,.xlsx,.mp4,.mp3,.avi,.wmv"
    >
      <Tooltip title={'文件格式支持 txt、HTML、Markdown、PDF、DOC、DOCX、XLS、XLSX、MP4、MP3、AVI、WMV，单个文件限制 50MB'}>
        <Button
          icon={
            <img
              src={uploadSvg}
              alt="上传"
            />
          }
        />
      </Tooltip>
    </Upload>
  );
}
