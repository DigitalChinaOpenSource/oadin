import { Form, Input, Modal } from 'antd';
import styles from './index.module.scss';
import { CloseOutlined } from '@ant-design/icons';
import { IUserType } from '@/pages/UserCenter/types';

interface IUserNameModalProps {
  visible: boolean;
  title?: string;
  userType: IUserType;
  onCancel: () => void;
  onConfirm: () => void;
}

const UserNameModal = ({ title, userType, onCancel, onConfirm, visible }: IUserNameModalProps) => {
  const [form] = Form.useForm();

  const handleOk = async () => {
    await form.validateFields();
    const formData = form.getFieldsValue();
    console.log('提交的表单数据:', formData);
    onConfirm();
    form.resetFields();
  };

  return (
    <Modal
      title={
        <div className={styles.headerTitle}>
          <span>{title || '用户名修改'}</span>
        </div>
      }
      open={visible}
      onCancel={onCancel}
      closeIcon={<CloseOutlined />}
      centered
      className={styles.authModal}
      width={480}
      maskClosable={true}
      destroyOnHidden={true}
      onOk={handleOk}
    >
      <Form
        form={form}
        name="userNameForm"
        autoComplete="off"
        className="loginForm"
      >
        <Form.Item
          name="userName"
          label={userType === 'person' ? '用户名' : '企业名称'}
          rules={[{ required: true, message: userType === 'person' ? '请输入用户名' : '请输入企业名称' }]}
        >
          <Input
            placeholder={userType === 'person' ? '请输入用户名' : '请输入企业名称'}
            allowClear
          />
        </Form.Item>
      </Form>
    </Modal>
  );
};

export default UserNameModal;
