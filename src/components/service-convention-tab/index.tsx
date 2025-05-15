import styles from './index.module.scss';

export default function serviceConvention() {
  return (
    <div className={styles.container}>
      <div className={styles.item}>
        <span>用户协议</span>
        <img
          // src={}
          alt="refresh"
        />
      </div>
      <div className={styles.item}>
        <span>隐私协议</span>
        <img
          // src={}
          alt="refresh"
        />
      </div>
    </div>
  );
}
