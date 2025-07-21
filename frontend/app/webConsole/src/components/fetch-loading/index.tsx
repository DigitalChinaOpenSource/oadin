import realLoadingSvg from '@/components/icons/real-loading.svg';
import styles from './index.module.scss';

export default function fetchLoading() {
  return (
    <div className={styles.fetchLoading}>
      <img
        src={realLoadingSvg}
        alt="loading"
      />
    </div>
  );
}
