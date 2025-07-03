import React, { useState } from 'react';
import { Button, Col, Row, Tooltip } from 'antd';
import ChatModelManage from './chat-model-manage';
import { PlusIcon, ClockCounterClockwiseIcon } from '@phosphor-icons/react';
import ChatHistoryDrawer from './chat-history-drawer';
import ChatView from './chat-view';
import useChatStore from './store/useChatStore';
import exchangeSvg from '@/components/icons/exchange.svg';
import useViewModel from './view-model';
import { ChooseModelDialog } from '@/components/choose-model-dialog';

import styles from './index.module.scss';

export default function ChatContainer() {
  const { historyVisible, setHistoryVisible, isLoading } = useChatStore();
  const [open, setOpen] = useState<boolean>(false);
  const vm = useViewModel();

  return (
    <div className={styles.chatContainer}>
      <div className={styles.header}>
        <div className={styles.title}>模型体验</div>
        <div className={styles.actions}>
          <Tooltip title={isLoading ? '对话中不可查看历史记录' : '对话历史'}>
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
                if (isLoading) return;
                setHistoryVisible(true);
              }}
            />
          </Tooltip>
          <Tooltip title={isLoading ? '对话中不可新建对话' : '新建对话'}>
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
                if (isLoading) return;
                vm.handleCreateNewChat();
              }}
            />
          </Tooltip>
          <Tooltip title={isLoading ? '对话中不可修改模型' : '切换模型后，将开启新会话'}>
            <div
              className={styles.changeModel}
              onClick={() => {
                if (isLoading) return;
                setOpen(true);
              }}
            >
              <img
                src={exchangeSvg}
                alt="切换模型图标"
              />
            </div>
          </Tooltip>
        </div>
      </div>
      <Row>
        <Col
          offset={5}
          md={14}
        >
          <div className={styles.chatContent}>
            <ChatModelManage />
            <ChatView isDownloadEmbed={vm.isDownloadEmbed} />
          </div>
        </Col>
      </Row>
      {historyVisible && (
        <ChatHistoryDrawer
          onHistoryDrawerClose={() => setHistoryVisible(false)}
          handleCreateNewChat={vm.handleCreateNewChat}
        />
      )}
      {open && (
        <ChooseModelDialog
          open={true}
          onCancel={() => setOpen(false)}
          fromWhere="chat-container"
        />
      )}
    </div>
  );
}
