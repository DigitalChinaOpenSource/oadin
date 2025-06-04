import { Button, Col, Row } from 'antd';
import createChat from '@/components/icons/createChat.svg';
import chatHistory from '@/components/icons/chatHistory.svg';
import ChatModelManage from './chat-model-manage';
import ChatHistoryDrawer from './chat-history-drawer';
import styles from './index.module.scss';
export default function ChatContainer() {
  return (
    <div className={styles.chatContainer}>
      <div className={styles.header}>
        <div className={styles.title}>模型体验</div>
        <div className={styles.actions}>
          <div
            className={styles.chatHistory}
            onClick={() => {
              console.log('查看对话历史');
            }}
          >
            <img
              src={chatHistory}
              alt="对话历史"
            />
          </div>
          <Button
            type="primary"
            className={styles.createChat}
            onClick={() => {
              console.log('创建新对话');
            }}
          >
            <img
              src={createChat}
              alt="创建新对话"
            />
            新建会话
          </Button>
        </div>
      </div>
      <Row>
        <Col
          offset={5}
          md={14}
        >
          <div className={styles.chatContent}>
            <ChatModelManage />
          </div>
        </Col>
      </Row>

      {/* <ChatHistoryDrawer /> */}
    </div>
  );
}
