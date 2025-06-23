import styles from './index.module.scss';
import { Button, Checkbox, Form, Input } from 'antd';
import useLoginStore from '@/store/loginStore.ts';

export interface IEnterpriseFormValues {
  email: string;
  password: string;
}
interface LoginFormProps {
  type?: 'login' | 'bind';
  onOk?: (values: IEnterpriseFormValues) => void;
  showAgreed?: boolean;
}

const LoginEnterprise: React.FC<LoginFormProps> = ({ onOk, showAgreed = true }) => {
  const [form] = Form.useForm<IEnterpriseFormValues>();
  const onFinish = (values: IEnterpriseFormValues) => {
    console.log('Received values of form: ', values);
    onOk?.(values);
  };
  const { setCurrentStep } = useLoginStore();
  return (
    <div className={styles.loginEnterprise}>
      {/* 头部标题区 */}
      <div className={styles.headerSection}>
        <div className={styles.tabContainer}>
          <div className={styles.tabItem}>
            <span className={`${styles.tabTitle} `}>邮箱登录</span>
          </div>
        </div>
      </div>
      <div className={styles.formContainer}>
        <Form
          form={form}
          name="emailForm"
          autoComplete="off"
          onFinish={onFinish}
          className={styles.loginForm}
        >
          <Form.Item
            name="email"
            validateTrigger={['onChange', 'onBlur']}
            rules={[
              {
                validator: (_: unknown, value: string) => {
                  if (!value) {
                    return Promise.reject(new Error('请输入邮箱地址'));
                  }
                  if (!/^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$/.test(value)) {
                    return Promise.reject(new Error('邮箱格式错误，请重新输入！'));
                  }
                  return Promise.resolve();
                },
                validateTrigger: ['onBlur'],
              },
            ]}
          >
            <Input
              style={{ width: '100%', height: '40px' }}
              placeholder="请输入邮箱地址"
            />
          </Form.Item>
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
              style={{ width: '100%', height: '40px' }}
              placeholder="请输入密码"
            />
          </Form.Item>
          {/* 同意协议复选框 */}
          {showAgreed && (
            <Form.Item
              name="agreed"
              valuePropName="checked"
              initialValue={false}
              rules={[
                {
                  validator: (_, value) => {
                    return value ? Promise.resolve() : Promise.reject(new Error('请阅读并同意用户协议和隐私政策'));
                  },
                },
              ]}
            >
              <Checkbox>
                我已阅读并同意 OADIN{' '}
                <a
                  href={'#'}
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  用户协议
                </a>{' '}
                和{' '}
                <a
                  href={'#'}
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  隐私政策
                </a>
              </Checkbox>
            </Form.Item>
          )}
          <Form.Item style={{ marginBottom: 0 }}>
            <Button
              htmlType="submit"
              type="primary"
              className={styles.loginButton}
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
