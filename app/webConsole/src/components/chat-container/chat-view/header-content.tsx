import { FolderIcon, XCircleIcon } from '@phosphor-icons/react';
import type { UploadProps, UploadFile } from 'antd';
import { message } from 'antd';
import { httpRequest } from '@/utils/httpRequest';
import useUploadFileListStore from '../store/useUploadFileListStore';
import rollingSvg from '@/components/icons/rolling.svg';

export const HeaderContent = () => {
  const { uploadFileList, setUploadFileList } = useUploadFileListStore();
  const handleRemove = async (file: UploadFile) => {
    try {
      const fileId = file.status === 'done' && file.response?.id;
      if (fileId) {
        await httpRequest.del('/playground/file', {
          file_id: fileId,
        });
      }

      // 无论是否调用接口，都从列表中移除
      setUploadFileList(uploadFileList.filter((f) => f.uid !== file.uid));
      // message.success(`已删除文件: ${file.name}`);

      return true;
    } catch (error: Error | any) {
      message.error(`删除文件失败: ${error?.message || '未知错误'}`);
      return false;
    }
  };

  if (uploadFileList.length === 0) return null;
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 16, flexWrap: 'wrap' }}>
      {uploadFileList.map((file) => (
        <div
          key={file.uid}
          className="upload-file-item"
        >
          {file.status === 'uploading' && (
            <div className="file-icon uploading-icon">
              <img
                src={rollingSvg}
                alt=""
              />
            </div>
          )}
          {file.status === 'error' && (
            <div className="file-icon error-icon">
              <XCircleIcon
                width={16}
                height={16}
                weight="fill"
                fill="#e85951"
              />
            </div>
          )}
          {file.status === 'done' && (
            <div className="file-icon done-icon">
              <FolderIcon
                width={16}
                height={16}
                fill="#ffffff"
              />
            </div>
          )}

          {file.name}
          <div
            className="upload-file-remove"
            onClick={(e) => {
              e.stopPropagation();
              handleRemove(file);
            }}
          >
            <XCircleIcon
              width={16}
              height={16}
              fill="#9ca3af"
              weight="fill"
            />
          </div>
        </div>
      ))}
    </div>
  );
};
