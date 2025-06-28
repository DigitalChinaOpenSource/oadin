import styles from '@/pages/UserCenter/accountSetting/sureAuthModal/index.module.scss';
import { CloseOutlined } from '@ant-design/icons';
import { Form, Modal, type UploadFile } from 'antd';
import ImageUpload from '@/pages/Login/components/ImageUpload';
import LoginEnterpriseIcon from '@/assets/login-enterprise-icon.svg';
import LoginFrontedIcon from '@/assets/login-fronted-icon.svg';
import LoginBackIcon from '@/assets/login-back-icon.svg';
import { IAccountInfo, IUserType } from '@/pages/UserCenter/types';

interface IAuthUploadModalProps {
  visible: boolean;
  onCancel: () => void;
  onConfirm: (formData: any) => void;
  userType: IUserType;
  title?: string;
  accountInfo: IAccountInfo; // 可能需要根据实际情况调整类型
}

const AuthUploadModal = ({ visible, onCancel, onConfirm, userType = 'person', title, accountInfo }: IAuthUploadModalProps) => {
  const [form] = Form.useForm();
  const isPerson = userType === 'person';

  const handleSure = async () => {
    await form.validateFields();
    console.log(form.getFieldsValue());
    onConfirm(form.getFieldsValue());
  };

  // 上传企业营业执照
  const renderEnterpriseForm = () => {
    return (
      <Form
        form={form}
        name={'authEnterpriseForm'}
        autoComplete={'off'}
        className="loginForm"
      >
        <ImageUpload
          height={'auto'}
          title={'上传营业执照'}
          bgIcon={LoginEnterpriseIcon}
          name="enterpriseIcon"
          rules={[{ required: true, message: '请上传营业执照' }]}
          onChange={(value) => form.setFieldsValue({ enterpriseIcon: value })}
          value={[{ url: accountInfo.licenseImageUrl || '', name: '' }] as UploadFile[]}
        />
      </Form>
    );
  };

  //上传个人身份证
  const renderPersonForm = () => {
    return (
      <Form
        form={form}
        name={'authPersonForm'}
        autoComplete={'off'}
        className="loginForm"
      >
        <ImageUpload
          height={'auto'}
          bgIcon={LoginFrontedIcon}
          title="上传人像面"
          name="frontImage"
          rules={[{ required: true, message: '请上传身份证人像面照片' }]}
          onChange={(value) => form.setFieldsValue({ frontImage: value })}
        />
        <ImageUpload
          height={'auto'}
          bgIcon={LoginBackIcon}
          title="上传国徽面"
          name="backImage"
          rules={[{ required: true, message: '请上传身份证国徽面照片' }]}
          onChange={(value) => form.setFieldsValue({ backImage: value })}
        />
      </Form>
    );
  };

  return (
    <Modal
      title={
        <div className={styles.headerTitle}>
          <span>{title || '实名认证'}</span>
        </div>
      }
      open={visible}
      onCancel={onCancel}
      closeIcon={<CloseOutlined />}
      centered
      className={styles.authModal}
      width={480}
      maskClosable={false}
      onOk={handleSure}
      destroyOnHidden={true}
    >
      <div className={styles.description}>{userType === 'person' ? '为了保障您的账号安全和数据互通，请完成实名认证' : '采集企业营业执照信息是为了进行资质核验与服务保障'}</div>

      <div className={styles.formContent}>{userType === 'person' ? renderPersonForm() : renderEnterpriseForm()}</div>
    </Modal>
  );
};

export default AuthUploadModal;
