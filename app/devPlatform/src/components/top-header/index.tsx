import styles from './index.module.scss';
import favicon from '@/assets/favicon.png';
import DefaultUserIcon from '@/assets/userIcon.svg';
import CompanyIcon from '@/assets/companyIcon.svg';
import useAuthStore from '@/store/authStore.ts';
import { Button } from 'antd';

export default function TopHeader() {
  const { user, changeUser } = useAuthStore();
  console.log(user);

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
          src={user.avatar || (user.type === 'person' ? DefaultUserIcon : CompanyIcon)}
          alt=""
        />
      </div>
    </div>
  );
}
