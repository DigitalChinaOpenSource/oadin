import styles from './index.module.scss';
type CardItemType = {
  logo: string;
  title: string;
  description: string;
  id: string;
};
type CardItemProps = {
  clientData: CardItemType;
};

const CardItem = (props: CardItemProps) => {
  const { logo, title, description } = props.clientData;
  return (
    <div className={styles.clientItem}>
      <div className={styles.clientIcon}>
        <img
          src={logo}
          alt={title}
        />
      </div>
      <div className={styles.clientContent}>
        <div className={styles.clientName}>{title}</div>
        <div className={styles.clientDesc}>{description}</div>
      </div>
    </div>
  );
};

export default function RecommendedClient() {
  const testList: CardItemType[] = [
    { id: '1', logo: '1', title: '测试1', description: '测试内容1' },
    { id: '2', logo: '2', title: '测试2', description: '测试内容2' },
    { id: '3', logo: '2', title: '测试2', description: '测试内容2' },
    { id: '4', logo: '2', title: '测试2', description: '测试内容2' },
    { id: '5', logo: '2', title: '测试2', description: '测试内容2' },
  ];
  return (
    <div className={styles.recommendedClient}>
      <div className={styles.recommendedHeader}>推荐客户端</div>
      {testList.map((item) => {
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
