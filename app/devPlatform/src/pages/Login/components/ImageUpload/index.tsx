import styles from './index.module.scss';
import { Upload, Image, Button, Form, Spin, App } from 'antd';
import { PlusOutlined, EyeOutlined, DeleteOutlined, ReloadOutlined } from '@ant-design/icons';
import { useState, useEffect } from 'react';
import type { UploadFile, UploadProps } from 'antd';
import { IImageUploadProps } from '@/pages/Login/types';
import { useOssSignStore } from '@/store/ossSignStore.ts';
import OSS from 'ali-oss';

const ImageUpload = ({ title = '上传图片', maxSize = 1, accept = ['image/jpeg', 'image/png'], height = 112, value, onChange, name, rules, bgIcon }: IImageUploadProps) => {
  const [fileList, setFileList] = useState<UploadFile[]>([]);
  const [previewOpen, setPreviewOpen] = useState(false);
  const [previewImage, setPreviewImage] = useState('');
  const { getOssSign, ossSign } = useOssSignStore();
  const { message } = App.useApp();

  const [loading, setLoading] = useState(false);

  const customRequest = async (options: any) => {
    console.log('customRequest file', 111111111111);

    const { file, onSuccess, onError } = options;
    console.log('customRequest file', options);
    setLoading(true);
    try {
      // 1. 获取 OSS 签名信息
      const ossSign = await getOssSign();
      if (!ossSign) {
        setLoading(false);
        return;
      }
      // 2. 初始化 OSS 客户端
      console.log(ossSign);
      const client = new OSS({
        region: ossSign.region,
        accessKeyId: ossSign.accessKeyId,
        bucket: ossSign.bucket,
        secure: true,
        accessKeySecret: ossSign.accessKeySecret,
        stsToken: ossSign.securityToken,
      });
      console.log('OSS 客户端初始化成功', client);
      const result = await client.put(`/${ossSign.dir}${file.name}`, file);
      // const res = await httpRequest.put(`/${ossSign.dir}${file.name}`, file, {
      //   headers: {
      //     'x-oss-policy': ossSign.policy,
      //     'x-oss-signature': ossSign.signature,
      //   },
      // });
      console.log('上传结果', result);

      // if (response.status === 200) {
      //   message.success('文件上传成功');
      //   onSuccess(response, file);
      //   handleSuccess(file);
      // } else {
      //   setLoading(false);
      //   message.error('文件上传失败');
      //   onError(new Error('上传失败'));
      // }
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

  const handleSuccess = (file: UploadFile) => {
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
                      fileList={fileList} // 修复：确保 fileList 正确传递
                      // onChange={handleChange}
                      beforeUpload={beforeUpload}
                      maxCount={1}
                      showUploadList={false}
                      multiple={false}
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
              </>
            )}
          </div>
        ) : (
          <Upload
            listType="picture-card"
            action=""
            fileList={fileList} // 修复：确保 fileList 正确传递
            // onChange={handleChange}
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
