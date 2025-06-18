/** MCP工具对话块 */
import { useState } from 'react';
import type { CollapseProps, TabsProps } from 'antd';
import { CaretRightOutlined } from '@ant-design/icons';
import { Collapse, Tabs, Tooltip } from 'antd';
import macpChatSvg from '@/components/icons/mcp-chat.svg';
import amap from '@/components/icons/amap.png';
import { CheckCircleIcon } from '@phosphor-icons/react';
import arrowUp from '@/components/icons/arrow-up.svg';
import arrowDown from '@/components/icons/arrow-down.svg';
import JsonParsePanel from '@/components/json-parse-panel';
import useChatStore from '../../store/useChatStore';
import styles from './index.module.scss';

interface IDataSourceItem {
  name: string;
  desc: string;
  logo: string;
  inputParams: string;
  outputParams: string;
  status: 'success' | 'error' | 'progress';
  executionTime?: number; // 执行时间
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

  // 计算已完成的工具调用
  const completedCalls = data.filter((item) => item.status === 'success' || item.status === 'error');
  const inProgressCalls = data.filter((item) => item.status === 'progress');

  // 计算总耗时
  const totalExecutionTime = data.reduce((sum, item) => sum + (item.executionTime || 0), 0);

  const jsonParsePannel = (content: string) => {
    return (
      <JsonParsePanel
        propsContentStyles={{ backgroundColor: 'unset' }}
        isConfig={false}
        maxHeight="240px"
        code={content}
      />
    );
  };

  const tabItems = (item: IDataSourceItem) => [
    {
      key: 'inputParams',
      label: '输入参数',
      children: jsonParsePannel(item.inputParams),
    },
    {
      key: 'outputParams',
      label: '输出参数',
      children: item.status === 'progress' ? <div className={styles.progressLoading}>工具调用中...</div> : jsonParsePannel(item.outputParams),
    },
  ];

  const inputoutParams = (item: IDataSourceItem) => {
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
              alt={item.name}
            />
          </div>
          <Tooltip title={item?.name}>
            <div className={styles.mcpTitle}>{item?.name}</div>
          </Tooltip>
        </div>
        <div className={styles.fill}></div>
        <Tooltip title={item?.desc}>
          <div className={styles.mcpDesc}>{item?.desc || '暂无描述信息'}</div>
        </Tooltip>
        <div className={styles.fill}></div>
        <div className={styles.mcpStatus}>
          {item?.status === 'success' && (
            <>
              <div className={styles.successDot}></div>
              <span className={styles.mcpSuccess}>成功</span>
            </>
          )}
          {item?.status === 'error' && (
            <>
              <div className={styles.errorDot}></div>
              <span className={styles.mcpError}>失败</span>
            </>
          )}
          {item?.status === 'progress' && (
            <>
              <div className={styles.progressDot}></div>
              <span className={styles.mcpProgress}>进行中</span>
            </>
          )}
        </div>
      </div>
    );
  };

  // 构建折叠面板项
  const collapseItems: CollapseProps['items'] = [
    // 显示已完成的调用
    ...completedCalls.map((item, index) => ({
      key: `completed-${index}`,
      label: mcpToolHeader(item),
      children: inputoutParams(item),
    })),
    // 显示进行中的调用
    ...inProgressCalls.map((item, index) => ({
      key: `progress-${index}`,
      label: mcpToolHeader(item),
      children: inputoutParams(item),
    })),
  ];

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
              <div className={styles.statusText}>工具调用已完成，共执行 {data.length} 次</div>
              <div className={styles.coastTime}>（总用时 {(totalExecutionTime / 1000).toFixed(1)} 秒）</div>
            </>
          )}
        </div>

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
      </div>

      <div className={`${styles.content} ${isExpanded ? styles.expanded : ''}`}>
        <Collapse
          items={collapseItems}
          bordered={false}
          expandIcon={({ isActive }) => <CaretRightOutlined rotate={isActive ? 90 : 0} />}
          // defaultActiveKey={collapseItems.map((item, index) => item?.key || index)}
        />
      </div>
    </div>
  );
}
