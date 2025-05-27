import styles from './index.module.scss';
import { useRecommendedClient } from '@/components/mcp-manage/mcp-detail/recommended-client/view.module.ts';
type CardItemType = {
  icon: string;
  name: string;
  description: string;
  id: string | number;
};
type CardItemProps = {
  clientData: CardItemType;
};

const CardItem = (props: CardItemProps) => {
  const { icon, name, description } = props.clientData;
  return (
    <div
      className={styles.clientItem}
      // onClick={() => window.open('test-app://')}
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
  const { clients } = useRecommendedClient();
  // const testList: CardItemType[] = [
  //   { id: '1', logo: '1', title: '测试1', description: '测试内容1' },
  //   { id: '2', logo: '2', title: '测试2', description: '测试内容2' },
  //   { id: '3', logo: '2', title: '测试2', description: '测试内容2' },
  //   { id: '4', logo: '2', title: '测试2', description: '测试内容2' },
  //   { id: '5', logo: '2', title: '测试2', description: '测试内容2' },
  // ];
  return (
    <div className={styles.recommendedClient}>
      <div className={styles.recommendedHeader}>推荐客户端</div>
      {clients.map((item: any) => {
        return (
          <CardItem
            clientData={item}
            key={item.id}
          />
        );
      })}
    </div>
  );
}
