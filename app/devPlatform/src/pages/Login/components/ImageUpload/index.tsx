import styles from './index.module.scss';
import { Upload, message, Image, Button, Form } from 'antd';
import { PlusOutlined, EyeOutlined, DeleteOutlined, ReloadOutlined } from '@ant-design/icons';
import { useState, useEffect } from 'react';
import type { UploadFile, UploadProps } from 'antd';

interface ImageUploadProps {
  title?: string;
  maxSize?: number; // 单位：MB
  accept?: string[];
  height?: number | string;
  value?: UploadFile[];
  onChange?: (fileList: UploadFile[]) => void;
  name?: string;
  rules?: any[];
  action?: string; // 添加上传地址
  customRequest?: (options: any) => void; // 添加自定义上传方法
}

const ImageUpload: React.FC<ImageUploadProps> = ({
  title = '上传图片',
  maxSize = 1,
  accept = ['image/jpeg', 'image/png'],
  height = 112,
  value,
  onChange,
  name,
  rules,
  action = '/api/upload', // 默认上传地址
  customRequest,
}) => {
  const [fileList, setFileList] = useState<UploadFile[]>(value || []);
  const [previewOpen, setPreviewOpen] = useState(false);
  const [previewImage, setPreviewImage] = useState('');

  // 同步外部表单值到内部状态
  useEffect(() => {
    if (value) {
      setFileList(value);
    }
  }, [value]);

  const handleChange: UploadProps['onChange'] = ({ fileList: newFileList }) => {
    setFileList(newFileList);
    console.log('newFileList', newFileList);
    // 将值传递给表单
    if (onChange) {
      onChange(newFileList);
    }
  };

  const beforeUpload = (file: File) => {
    const isAcceptType = accept.includes(file.type);
    if (!isAcceptType) {
      message.error(`只能上传 ${accept.map((type) => type.split('/')[1].toUpperCase()).join('/')} 格式的图片!`);
    }
    const isLtMaxSize = file.size / 1024 / 1024 < maxSize;
    if (!isLtMaxSize) {
      message.error(`图片大小不能超过 ${maxSize}MB!`);
    }
    return isAcceptType && isLtMaxSize;
  };

  const handlePreview = () => {
    if (fileList.length > 0) {
      setPreviewImage(fileList[0].thumbUrl || fileList[0].url || '');
      setPreviewOpen(true);
    }
  };

  const handleDelete = () => {
    const newFileList: UploadFile[] = [];
    setFileList(newFileList);
    // 将空数组传递给表单
    if (onChange) {
      onChange(newFileList);
    }
  };

  const uploadComponent = (
    <>
      <div
        className={styles.uploadContainer}
        style={{ height }}
      >
        {fileList.length > 0 ? (
          <div
            className={styles.previewContainer}
            style={{ height }}
          >
            <Image
              src={fileList[0].thumbUrl || fileList[0].url}
              alt={title}
              className={styles.previewImage}
              preview={false}
            />
            <div className={styles.previewOverlay}>
              <div className={styles.actionButtons}>
                <Button
                  icon={<EyeOutlined />}
                  className={styles.actionButton}
                  onClick={handlePreview}
                >
                  预览
                </Button>
                <Button
                  icon={<DeleteOutlined />}
                  className={styles.actionButton}
                  onClick={handleDelete}
                >
                  删除
                </Button>
                <Upload
                  listType="text"
                  fileList={[]}
                  onChange={handleChange}
                  beforeUpload={beforeUpload}
                  maxCount={1}
                  showUploadList={false}
                  multiple={false}
                  action={action} // 添加上传地址
                  customRequest={customRequest} // 添加自定义上传方法
                >
                  <Button
                    icon={<ReloadOutlined />}
                    className={styles.actionButton}
                  >
                    重新上传
                  </Button>
                </Upload>
              </div>
            </div>
          </div>
        ) : (
          <Upload
            listType="picture-card"
            fileList={[]}
            onChange={handleChange}
            beforeUpload={beforeUpload}
            maxCount={1}
            className={styles.idCardUpload}
            showUploadList={false}
            multiple={false}
            action={action} // 添加上传地址
            customRequest={customRequest} // 添加自定义上传方法
          >
            <div className={styles.uploadButton}>
              <PlusOutlined />
              <div className={styles.uploadText}>{title}</div>
            </div>
          </Upload>
        )}
      </div>
      {/* 图片预览弹窗 */}
      {previewOpen && (
        <Image
          width={0}
          style={{ display: 'none' }}
          preview={{
            visible: previewOpen,
            onVisibleChange: (visible) => setPreviewOpen(visible),
            src: previewImage,
          }}
        />
      )}
    </>
  );

  // 如果提供了name和rules，则用Form.Item包装
  if (name) {
    return (
      <Form.Item
        name={name}
        rules={rules}
        valuePropName="value"
        // getValueFromEvent 用于自定义从事件中获取表单值的逻辑
        // 在上传组件中，它负责从 Upload 组件的 onChange 事件中提取正确的值存入 Form
        // e 参数可能Upload 的 onChange 事件对象 (包含 fileList)，也可能是直接传入的 fileList 数组
        // 这里确保无论是哪种情况，都能正确提取 fileList 并存入表单项
        getValueFromEvent={(e) => e?.fileList || e}
      >
        {uploadComponent}
      </Form.Item>
    );
  }

  // 否则直接返回组件
  return uploadComponent;
};

export default ImageUpload;
