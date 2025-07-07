import styles from './index.module.scss';
import { useRecommendedClient } from '@/components/mcp-manage/mcp-detail/recommended-client/view-model';
import { CardItemProps } from '@/components/mcp-manage/mcp-detail/recommended-client/type.ts';

const CardItem = (props: CardItemProps) => {
  const { icon, name, description } = props.clientData;
  const { handleClick } = props;

  return (
    <div
      className={styles.clientItem}
      onClick={() => handleClick('test-app', props.clientData)}
    >
      <div className={styles.clientIcon}>
        <img
          src={icon}
          alt=""
        />
      </div>
      <div className={styles.clientContent}>
        <div className={styles.clientName}>{name}</div>
        <div className={styles.clientDesc}>{description}</div>
      </div>
    </div>
  );
};

export default function RecommendedClient() {
  const { clients, handleClick } = useRecommendedClient();

  return (
    <div className={styles.recommendedClient}>
      <div className={styles.recommendedHeader}>推荐客户端</div>
      {clients.map((item: any) => {
        return (
          <CardItem
            clientData={item}
            key={item.id}
            handleClick={handleClick}
          />
        );
      })}
    </div>
  );
}
