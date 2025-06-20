import styles from './index.module.scss';

const LoginEnterprise: React.FC = () => {
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
    </div>
  );
};
