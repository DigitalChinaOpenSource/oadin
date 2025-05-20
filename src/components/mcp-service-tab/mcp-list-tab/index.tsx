import Styles from './index.module.scss';
import { useNavigate } from 'react-router-dom';
type cardType = {
  title: string;
  content: string;
  icon?: string;
  tags: string[];
  serviceId: string;
};
export default function McpListTab() {
  const testList: cardType[] = [
    { serviceId: '1', title: '测试1', content: '测试内容1', icon: '', tags: ['标签1', '标签2'] },
    { serviceId: '2', title: '测试2', content: '测试内容2', icon: '', tags: ['标签3', '标签4'] },
  ];
  const navigate = useNavigate();
  const handelClick = (serviceId: string) => {
    navigate(`/mcp-service-detail?serviceId=${serviceId}&mcpFrom=mcpList`);
  };
  return (
    <div className={Styles.mcpManageList}>
      <div className={Styles.mcpManageListContent}>
        {testList.map((item) => (
          <div
            key={item.serviceId}
            className={Styles.card}
            onClick={() => handelClick(item.serviceId)}
          >
            <div className={Styles.cardTitle}>{item.title}</div>
            <div className={Styles.cardContent}>{item.content}</div>
            <div className={Styles.cardTags}>
              {item.tags.map((tag) => (
                <span
                  key={tag}
                  className={Styles.tag}
                >
                  {tag}
                </span>
              ))}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
