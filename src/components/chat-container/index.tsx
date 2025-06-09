import { Button, Col, Row, Tooltip } from 'antd';
import ChatModelManage from './chat-model-manage';
import { PlusIcon, ClockCounterClockwiseIcon } from '@phosphor-icons/react';
import ChatHistoryDrawer from './chat-history-drawer';
import ChatView from './chat-view';
import useViewModel from './useViewModel';
import styles from './index.module.scss';
export default function ChatContainer() {
  const { historyVisible, setHistoryVisible } = useViewModel();

  return (
    <div className={styles.chatContainer}>
      <div className={styles.header}>
        <div className={styles.title}>模型体验</div>
        <div className={styles.actions}>
          <Tooltip title={'对话历史'}>
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
                setHistoryVisible(true);
              }}
            />
          </Tooltip>
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
          </div>
        </Col>
      </Row>
      {historyVisible && <ChatHistoryDrawer onHistoryDrawerClose={() => setHistoryVisible(false)} />}
    </div>
  );
}
