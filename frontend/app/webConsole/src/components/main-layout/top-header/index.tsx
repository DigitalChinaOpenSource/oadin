import styles from './index.module.scss';
import favicon from '@/assets/favicon.png';

export default function TopHeader() {
  return (
    <div className={styles.topHeader}>
      <div className={styles.headerLeft}>
        <div className={styles.project}>
          <img
            src={favicon}
            alt=""
          />
          <div>OADIN</div>
        </div>
      </div>
      <div className={styles.personIcon}>
        <img
          src={favicon}
          alt=""
        />
      </div>
    </div>
  );
}
