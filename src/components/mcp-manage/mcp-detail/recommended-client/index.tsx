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

  const checkProtocol = (protocol: string) => {
    return new Promise((resolve) => {
      const timeout = 2000; // 2秒超时检测
      let timer: any = null;

      // 监听窗口是否失去焦点（如果跳转成功，页面会失去焦点）
      const onBlur = () => {
        clearTimeout(timer);
        window.removeEventListener('blur', onBlur);
        resolve(true); // 协议可用
      };

      window.addEventListener('blur', onBlur);

      // 尝试打开协议
      window.location.href = `${protocol}://`; // 或 window.open()

      // 如果 2 秒后未跳转，则认为协议未注册
      timer = setTimeout(() => {
        window.removeEventListener('blur', onBlur);
        resolve(false); // 协议不可用
      }, timeout);
    });
  };

  const handleClick = (protocol: string) => {
    if (!protocol) return;
    checkProtocol(protocol).then((isAvailable) => {
      if (isAvailable) {
        window.open(`${protocol}://`);
      } else {
        alert('未检测到应用，请安装后重试。');
      }
    });
  };

  return (
    <div
      className={styles.clientItem}
      onClick={() => handleClick('test-app')}
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
