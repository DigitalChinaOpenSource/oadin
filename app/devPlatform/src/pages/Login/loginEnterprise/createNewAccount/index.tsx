import styles from './index.module.scss';
import useLoginStore from '@/store/loginStore.ts';
import { ArrowLeftOutlined } from '@ant-design/icons';
const CreateNewAccount = () => {
  const { setCurrentStep } = useLoginStore();

  return (
    <div className={styles.createNewAccount}>
      <div
        className="goBack"
        onClick={() => setCurrentStep('enterpriseAccount')}
      >
        <ArrowLeftOutlined size={24} />
        <div>返回</div>
      </div>
      <div className="header">企业邮箱注册</div>
      <div className="formContainer"></div>
    </div>
  );
};

export default CreateNewAccount;
