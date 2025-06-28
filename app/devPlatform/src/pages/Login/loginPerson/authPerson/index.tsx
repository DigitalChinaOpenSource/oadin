import styles from './index.module.scss';
import ImageUpload from '@/pages/Login/components/ImageUpload';
import { Form, Button, App } from 'antd';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import LoginFrontedIcon from '@/assets/login-fronted-icon.svg';
import LoginBackIcon from '@/assets/login-back-icon.svg';

const AuthPerson = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const { message } = App.useApp();

  const handleSubmit = async () => {
    try {
      setLoading(true);
      const values = await form.validateFields();
      console.log('提交的表单数据：', values);

      // 这里可以看到表单中包含了上传的图片文件信息
      // values.frontImage 和 values.backImage 包含了上传的文件信息

      // 实际表单提交逻辑...
      message.success('提交成功！');
    } catch (error) {
      console.error('表单验证失败：', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className={styles.authPerson}>
      <div className="headerSection">
        <div className="tabContainer">
          <div className="tabItem">
            <span className="tabTitle">实名认证</span>
          </div>
        </div>
        <span className="subTitle">为了保障您的账号安全和数据互通，请完成实名认证</span>
      </div>

      <div className={styles.authContent}>
        <Form
          form={form}
          layout="vertical"
        >
          {/* 上传人像面区域 */}
          <ImageUpload
            bgIcon={LoginFrontedIcon}
            title="上传人像面"
            name="frontImage"
            rules={[{ required: true, message: '请上传身份证人像面照片' }]}
            onChange={(value) => form.setFieldsValue({ frontImage: value })} // 清除之前的图片
          />

          {/* 上传国徽面区域 */}
          <ImageUpload
            bgIcon={LoginBackIcon}
            title="上传国徽面"
            name="backImage"
            rules={[{ required: true, message: '请上传身份证国徽面照片' }]}
            onChange={(value) => form.setFieldsValue({ backImage: value })}
          />
          <div className="authOperate">
            <Button
              type="default"
              className="operateBtn"
              onClick={() => navigate('/')}
            >
              跳过此步骤
            </Button>
            <Button
              type="primary"
              className="sureBtn"
              onClick={handleSubmit}
              loading={loading}
            >
              确认认证
            </Button>
          </div>
        </Form>
      </div>
    </div>
  );
};

export default AuthPerson;
