/** MCP工具对话块 */
import { useState } from 'react';
import type { CollapseProps, TabsProps } from 'antd';
import { CaretRightOutlined } from '@ant-design/icons';
import { Collapse, Tabs } from 'antd';
import macpChatSvg from '@/components/icons/mcp-chat.svg';
import amap from '@/components/icons/amap.png';
import { CheckCircleIcon } from '@phosphor-icons/react';
import arrowUp from '@/components/icons/arrow-up.svg';
import arrowDown from '@/components/icons/arrow-down.svg';
import JsonParsePanel from '@/components/json-parse-panel';
import styles from './index.module.scss';

interface IDataSourceItem {
  name: string;
  desc: string;
  logo: string;
  inputParams: string;
  outputParams: string;
  status: 'success' | 'error' | 'progress';
}
interface IMcpToolChatData {
  dataSource: {
    data: IDataSourceItem[];
    status?: 'success' | 'error' | 'progress';
  };
}

export default function McpToolChat(props: IMcpToolChatData) {
  console.log('McpToolChat dataSource=====>', props);
  const { data, status } = props.dataSource;
  const [isExpanded, setIsExpanded] = useState(false);

  const jsonParsePannel = (data: string) => {
    return (
      <JsonParsePanel
        propsContentStyles={{ backgroundColor: 'unset' }}
        isConfig={false}
        maxHeight="240px"
        code={data}
      />
    );
  };

  const tabItems = (item: IDataSourceItem) => [
    {
      key: '1',
      label: '输入参数',
      children: jsonParsePannel(item.inputParams),
    },
    {
      key: '2',
      label: '输出参数',
      children: jsonParsePannel(item.outputParams),
    },
  ];

  const inputoutParams = (item: any) => {
    return (
      <div className={styles.inputoutParams}>
        <Tabs
          defaultActiveKey="inputParams"
          items={tabItems(item)}
          className={styles.inputoutTabs}
        />
      </div>
    );
  };

  const mcpToolHeader = (item: IDataSourceItem) => {
    return (
      <div className={styles.mcpToolHeader}>
        <div className={styles.mcpLeft}>
          <div className={styles.mcpLogo}>
            <img
              src={amap}
              alt="高德地图"
            />
          </div>
          <div className={styles.mcpTitle}>{item?.name}</div>
        </div>
        <div className={styles.fill}></div>
        <div className={styles.mcpDesc}>{item?.desc || '暂无描述信息'}</div>
        <div className={styles.fill}></div>
        <div className={styles.mcpStatus}>
          <div className={styles.successDot}></div>
          <span className={styles.mcpSuccess}>成功</span>
          {/* <span className={styles.mcpError}>失败</span> */}
        </div>
      </div>
    );
  };

  const collapseItems: CollapseProps['items'] = data.map((item: any, index: number) => ({
    key: `${index + 1}`,
    label: mcpToolHeader(item),
    children: inputoutParams(item),
  }));

  return (
    <div className={styles.mcpToolChat}>
      <div className={styles.header}>
        <div className={styles.chatStatus}>
          {status === 'progress' && (
            <>
              <img
                src={macpChatSvg}
                alt="工具调用中..."
              />
              <div className={styles.thinkingText}>工具调用中...</div>
            </>
          )}

          {status === 'error' && (
            <>
              <img
                src={macpChatSvg}
                alt="调用已停止"
              />
              <div className={styles.statusText}>调用已停止</div>
            </>
          )}
          {status === 'success' && (
            <>
              <CheckCircleIcon
                width={16}
                height={16}
                fill="#4f4dff"
              />
              <div className={styles.statusText}>工具调用已完成，共执行 3 次</div>
              <div className={styles.coastTime}>（用时 25 秒）</div>
            </>
          )}
        </div>

        {status !== 'progress' && (
          <div
            className={styles.collapse}
            onClick={() => setIsExpanded(!isExpanded)}
          >
            {isExpanded ? (
              <img
                src={arrowUp}
                alt="收起"
              />
            ) : (
              <img
                src={arrowDown}
                alt="展开"
              />
            )}
          </div>
        )}
      </div>
      <div className={`${styles.content} ${isExpanded ? styles.expanded : ''}`}>
        <Collapse
          items={collapseItems}
          bordered={false}
          expandIcon={({ isActive }) => <CaretRightOutlined rotate={isActive ? 90 : 0} />}
          defaultActiveKey={['1']}
        />
      </div>
    </div>
  );
}
