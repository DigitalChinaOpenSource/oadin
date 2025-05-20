import Styles from './index.module.scss';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { ArrowLeftOutlined } from '@ant-design/icons';

export default function McpDetail() {
  const [searchParams] = useSearchParams();
  const navicate = useNavigate();
  const serviceId = searchParams.get('serviceId');
  const mcpFrom = searchParams.get('mcpFrom');
  const handledGoBack = (): void => {
    // navicate(-1);
    navicate(`/mcp-service?mcpFrom=${mcpFrom}`);
  };
  return (
    <div className={Styles.mcpDetail}>
      <div
        className={Styles.goBack}
        onClick={handledGoBack}
      >
        <ArrowLeftOutlined className={Styles.backIcon} />
        <span className={Styles.backText}>返回</span>
      </div>
      <div className={Styles.detailTop}>
        <div className={Styles.card}>
          <div>概览标题</div>
          <div>概览标题</div>
          <div>概览标题</div>
          <div>概览标题</div>
          <div>概览标题</div>
          <div>概览标题</div>
          <div>概览标题</div>
        </div>
        <div className={Styles.topRight}>添加</div>
      </div>
      {/*分割线*/}
      <div className={Styles.Line}></div>
      <div className={Styles.detailContent}>detailContent</div>
    </div>
  );
}
