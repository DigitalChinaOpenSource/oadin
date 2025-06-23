import styles from './index.module.scss';
import AuthIcon from '@/assets/authIcon.svg';
import { Button, Form } from 'antd';
import ImageUpload from '@/pages/Login/components/ImageUpload';
import { useNavigate } from 'react-router-dom';
import { useState } from 'react';

const AuthEnterprise = () => {
  const [form] = Form.useForm();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const handleSubmit = () => {};

  return (
    <div className={styles.authEnterprise}>
      <div className="headerSection">
        <div className="tabContainer">
          <div className="tabItem">
            <span className="tabTitle">实名认证</span>
          </div>
        </div>
        <span className="subTitle">采集企业营业执照信息是为了进行资质校验与服务保障</span>
      </div>
      <div className={styles.authContent}>
        <div className={styles.desc}>
          <img
            src={AuthIcon}
            alt=""
          />
          <div>信息补充</div>
        </div>
        <Form
          form={form}
          name={'authEnterpriseForm'}
          autoComplete={'off'}
          className="loginForm"
        >
          <ImageUpload
            title={'上传营业执照'}
            name="enterpriseIcon"
            rules={[{ required: true, message: '请上传营业执照' }]}
            onChange={(value) => form.setFieldsValue({ enterpriseIcon: value })}
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

export default AuthEnterprise;
