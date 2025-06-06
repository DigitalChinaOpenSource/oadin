import { Button, Col, Row } from 'antd';
import ChatModelManage from './chat-model-manage';
import DeepThinkChat from './chat-components/deep-think-chat';
import McpToolChat from './chat-components/mcp-tool-chat';
import styles from './index.module.scss';
import { PlusIcon, ClockCounterClockwiseIcon } from '@phosphor-icons/react';
import { SelectMcp } from '@/components/select-mcp';
import ChatView from './chat-view';
export default function ChatContainer() {
  return (
    <div className={styles.chatContainer}>
      <div className={styles.header}>
        <div className={styles.title}>模型体验</div>
        <div className={styles.actions}>
          <Button
            type="text"
            className={styles.chatHistory}
            icon={
              <ClockCounterClockwiseIcon
                size={16}
                fill="#27272a"
              />
            }
            onClick={() => {
              console.log('查看对话历史');
            }}
          />
          <Button
            type="text"
            className={styles.createChat}
            icon={
              <PlusIcon
                size={16}
                fill="#27272a"
              />
            }
            onClick={() => {
              console.log('创建新对话');
            }}
          />
        </div>
      </div>
      <Row>
        <Col
          offset={5}
          md={14}
        >
          <div className={styles.chatContent}>
            <ChatModelManage />
            <ChatView />
            {/* <DeepThinkChat />
            <McpToolChat /> */}
          </div>
        </Col>
      </Row>
      {/* <ChatHistoryDrawer /> */}
    </div>
  );
}
