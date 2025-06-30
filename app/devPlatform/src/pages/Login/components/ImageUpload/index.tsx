import styles from './index.module.scss';
import { Upload, Image, Button, Form, Spin, App } from 'antd';
import { PlusOutlined, EyeOutlined, DeleteOutlined, ReloadOutlined } from '@ant-design/icons';
import { useState, useEffect } from 'react';
import { IImageUploadProps } from '@/pages/Login/types';
import { useUserCenterView } from '@/pages/UserCenter/useUserCenterView.ts';

const ImageUpload = ({ title = '上传图片', maxSize = 1, accept = ['image/jpeg', 'image/png'], height = 112, value, onChange, name, rules, bgIcon }: IImageUploadProps) => {
  const [fileList, setFileList] = useState<any[]>([]);
  const [previewOpen, setPreviewOpen] = useState(false);
  const [previewImage, setPreviewImage] = useState('');
  const { message } = App.useApp();
  const { uploadRealNameAuthPhoto } = useUserCenterView();

  const [loading, setLoading] = useState(false);

  const customRequest = async (options: any) => {
    console.log('customRequest file', 111111111111);

    const { file } = options;
    console.log('customRequest file', options);
    setLoading(true);
    try {
      const formData = new FormData();
      formData.append('file', file);
      formData.append('name', file.name);
      const res = await uploadRealNameAuthPhoto(formData);

      console.log('上传结果', res);
      if (res.code === 200) {
        setFileList([{ url: res.data.licenseImageUrl }]);
        handleSuccess({ url: res.data.licenseImageUrl });
        setLoading(false);
      }
    } catch (error) {
      console.log('上传失败', error);
      setLoading(false);
    }
  };

  // 同步外部表单值到内部状态
  useEffect(() => {
    console.log('value changed', value);
    if (value && value.length && value[0].url) {
      console.log('value changed22222', value);
      handleSuccess(value[0]);
    }
  }, [value]);

  const handleSuccess = (file: { url: string }) => {
    setFileList([file]);
    if (onChange) {
      onChange([file]);
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
    console.log(isAcceptType && isLtMaxSize);
    return isAcceptType && isLtMaxSize;
  };

  const handlePreview = () => {
    if (fileList.length > 0) {
      setPreviewImage(fileList[0].thumbUrl || fileList[0].url || '');
      setPreviewOpen(true);
    }
  };

  const handleDelete = () => {
    setFileList([]);
    // 将空数组传递给表单
    if (onChange) {
      onChange([]);
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
            {loading ? (
              <Spin />
            ) : (
              <>
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
                      action=""
                      fileList={[]} // 修复：确保 fileList 正确传递
                      beforeUpload={beforeUpload}
                      maxCount={1}
                      showUploadList={false}
                      multiple={false}
                      customRequest={customRequest} // 添加自定义上传方法
                      disabled={loading} // 上传时禁用上传按钮
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
              </>
            )}
          </div>
        ) : (
          <Upload
            listType="picture-card"
            action=""
            fileList={[]} // 修复：确保 fileList 正确传递
            beforeUpload={beforeUpload}
            maxCount={1}
            className={styles.idCardUpload}
            showUploadList={false}
            multiple={false}
            customRequest={customRequest} // 添加自定义上传方法
            disabled={loading} // 上传时禁用上传按钮
          >
            <div className={styles.uploadButton}>
              {loading ? (
                <Spin />
              ) : bgIcon ? (
                <img
                  src={bgIcon || ''}
                  alt=""
                />
              ) : (
                <>
                  <PlusOutlined />
                  <div className={styles.uploadText}>{title}</div>
                </>
              )}
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
