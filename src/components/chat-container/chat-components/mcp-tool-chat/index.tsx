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

// interface IMcpToolChatProps {}

export default function McpToolChat() {
  // TODO 思考完毕之后自动收起
  const [isExpanded, setIsExpanded] = useState(false);

  const jsonParsePannel = () => {
    return (
      <JsonParsePanel
        propsContentStyles={{ backgroundColor: 'unset' }}
        isConfig={false}
        maxHeight="240px"
        code={JSON.stringify(
          {
            content: [
              { type: 'text', text: 'Geocoding failed: USERKEY_PLAT_NOMATCH' },
              { type: 'text', text: 'Geocoding failed: USERKEY_PLAT_NOMATCH' },
              { type: 'text', text: 'Geocoding failed: USERKEY_PLAT_NOMATCH' },
              { type: 'text', text: 'Geocoding failed: USERKEY_PLAT_NOMATCH' },
            ],
            isError: true,
          },
          null,
          2,
        )}
      />
    );
  };

  const tabItems: TabsProps['items'] = [
    {
      key: 'inputParams',
      label: '输入参数',
      children: jsonParsePannel(),
    },
    {
      key: 'outputParams',
      label: '输出参数',
      children: jsonParsePannel(),
    },
  ];

  const inputoutParams = () => {
    return (
      <div className={styles.inputoutParams}>
        <Tabs
          defaultActiveKey="inputParams"
          items={tabItems}
          className={styles.inputoutTabs}
        />
      </div>
    );
  };

  const mcpToolHeader = () => {
    return (
      <div className={styles.mcpToolHeader}>
        <div className={styles.mcpLeft}>
          <div className={styles.mcpLogo}>
            <img
              src={amap}
              alt="高德地图"
            />
          </div>
          <div className={styles.mcpTitle}>maps_anyTool</div>
        </div>
        <div className={styles.fill}></div>
        <div className={styles.mcpDesc}>
          获取北京经纬度坐标获取北京经纬度坐标获取北京经纬度坐标获取北京经纬度坐标获取北京经纬度坐标获取北京经纬度坐标获取北京经纬度坐标获取北京经纬度坐标获取北京经纬度坐标
        </div>
        <div className={styles.fill}></div>
        <div className={styles.mcpStatus}>
          <div className={styles.successDot}></div>
          <span className={styles.mcpSuccess}>成功</span>
          {/* <span className={styles.mcpError}>失败</span> */}
        </div>
      </div>
    );
  };

  const items: CollapseProps['items'] = [
    {
      key: '1',
      label: mcpToolHeader(),
      children: inputoutParams(),
    },
    {
      key: '2',
      label: 'This is panel header 2',
      children: inputoutParams(),
    },
    {
      key: '3',
      label: 'This is panel header 3',
      children: inputoutParams(),
    },
  ];

  return (
    <div className={styles.mcpToolChat}>
      <div className={styles.header}>
        <div className={styles.chatStatus}>
          {/* <>
            <img
              src={macpChatSvg}
              alt="工具调用中..."
            />
            <div className={styles.thinkingText}>工具调用中...</div>
          </> */}
          {/* <>
            <img
              src={macpChatSvg}
              alt="调用已停止"
            />
            <div className={styles.statusText}>调用已停止</div>
          </> */}
          <>
            <CheckCircleIcon
              width={16}
              height={16}
              fill="#4f4dff"
            />
            <div className={styles.statusText}>工具调用已完成，共执行 3 次</div>
            <div className={styles.coastTime}>（用时 25 秒）</div>
          </>
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
          items={items}
          bordered={false}
          expandIcon={({ isActive }) => <CaretRightOutlined rotate={isActive ? 90 : 0} />}
          defaultActiveKey={['1']}
        />
      </div>
    </div>
  );
}
