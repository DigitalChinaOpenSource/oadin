import styles from './index.module.scss';
import { App, Button, Checkbox, Form, Input, message } from 'antd';
import useLoginStore from '@/store/loginStore.ts';
import AgreedCheckBox from '@/pages/Login/components/agreedCheckBox';
import EmailInput from '@/pages/Login/components/emailInput';
import { useLoginView } from '@/pages/Login/useLoginView.ts';
import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import useAuthStore from '@/store/authStore.ts';

export interface IEnterpriseFormValues {
  email: string;
  password: string;
}
interface LoginFormProps {
  type?: 'login' | 'bind';
  onOk?: (values: IEnterpriseFormValues) => void;
  showAgreed?: boolean;
}

const LoginEnterprise: React.FC<LoginFormProps> = ({ showAgreed = true }) => {
  const [form] = Form.useForm<IEnterpriseFormValues>();
  const { loginWithEnterprise, getEnterpriseInfo } = useLoginView();
  const { enterpriseAccountData, setEnterpriseAccountData, setCurrentStep } = useLoginStore();
  const navigate = useNavigate();
  const { login } = useAuthStore();
  const { message } = App.useApp();

  const onFinish = async (values: IEnterpriseFormValues) => {
    console.log('Received values of form: ', values);
    const res = await loginWithEnterprise({ email: values.email, password: values.password });
    console.log('企业登录结果', res);
    if (res.code === 200) {
      // 登录必定不是新用户 直接进入
      login(res.data.user, res.data.token);
      navigate('/');
    } else if (res.code === 401) {
      // 登录失败
      form.resetFields(['password']);
      setEnterpriseAccountData({ ...values, password: '' });
      form.setFields([{ name: 'password', errors: [res.message || '登录失败，请重试'] }]);
    } else {
      message.error(res.message || '登录失败，请重试');
    }
  };

  useEffect(() => {
    // if (enterpriseAccountData) {
    //   form.setFieldsValue(enterpriseAccountData);
    // }
  }, []);

  return (
    <div className={styles.loginEnterprise}>
      {/* 头部标题区 */}
      <div className="headerSection">
        <div className="tabContainer">
          <div className="tabItem">
            <span className="tabTitle">邮箱登录</span>
          </div>
        </div>
      </div>
      <div className={styles.formContainer}>
        <Form
          form={form}
          name="emailForm"
          autoComplete="off"
          onFinish={onFinish}
          className="loginForm"
          onValuesChange={(values: IEnterpriseFormValues) => {
            setEnterpriseAccountData(values);
          }}
        >
          <EmailInput
            form={form}
            codeFiled={'email'}
          />
          <Form.Item
            name="password"
            validateTrigger={['onBlur']}
            rules={[
              {
                validator: (_: unknown, value: string) => {
                  if (!value) {
                    return Promise.reject(new Error('请输入密码'));
                  }
                  return Promise.resolve();
                },
              },
            ]}
          >
            <Input.Password
              className="formInput"
              placeholder="请输入密码"
              allowClear
            />
          </Form.Item>
          {/* 同意协议复选框 */}
          {showAgreed && (
            <AgreedCheckBox
              form={form}
              codeFiled={'agreed'}
            />
          )}
          <Form.Item style={{ marginBottom: 0 }}>
            <Button
              htmlType="submit"
              type="primary"
              className="loginButton"
            >
              登录
            </Button>
          </Form.Item>
        </Form>
      </div>
      <div
        className={styles.forgetPassword}
        onClick={() => setCurrentStep('forgetPassword')}
      >
        忘记密码
      </div>
      <div className={styles.createNewAccount}>
        <span>没有账号？</span>
        <span
          className={styles.createDesc}
          onClick={() => setCurrentStep('createAccount')}
        >
          创建一个账号
        </span>
      </div>
    </div>
  );
};

export default LoginEnterprise;
