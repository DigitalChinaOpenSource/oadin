import Styles from './index.module.scss';
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
    <div className={Styles.clientItem}>
      <div className={Styles.clientIcon}>
        <img
          src={logo}
          alt=""
        />
      </div>
      <div className={Styles.clientContent}>
        <div className={Styles.clientName}>{title}</div>
        <div className={Styles.clientDesc}>{description}</div>
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
    <div className={Styles.recommendedClient}>
      <div className={Styles.recommendedHeader}>推荐客户端</div>
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
