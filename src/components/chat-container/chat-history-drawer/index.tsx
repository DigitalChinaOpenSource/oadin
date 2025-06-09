import { Drawer, Button, Tooltip, Popconfirm } from 'antd';
import { useChatHistoryDrawer } from '@/components/chat-container/chat-history-drawer/view.module.ts';
import { useEffect, useMemo, useState } from 'react';
import styles from './index.module.scss';
import dayjs from 'dayjs';
import { IChatHistoryItem } from '@/components/chat-container/chat-history-drawer/types.ts';
import { TrashIcon } from '@phosphor-icons/react';

export interface IChatHistoryDrawerProps {
  onHistoryDrawerClose?: () => void;
}

// 定义分组结果接口
interface GroupedChatHistory {
  today: IChatHistoryItem[];
  yesterday: IChatHistoryItem[];
  last7Days: IChatHistoryItem[];
  earlier: IChatHistoryItem[];
}

/**
 * 按日期对聊天历史记录进行分组
 * @param history 聊天历史记录数组
 * @returns 分组后的聊天历史记录对象
 */
function groupChatHistoryByDate(history: IChatHistoryItem[]): GroupedChatHistory {
  const now = dayjs();
  const todayStart = now.startOf('day');
  const yesterdayStart = now.subtract(1, 'day').startOf('day');
  const last7DaysStart = now.subtract(7, 'day').startOf('day');

  return history.reduce<GroupedChatHistory>(
    (groups, item) => {
      const itemDate = dayjs(item.createdAt);

      if (itemDate.isSame(todayStart, 'day')) {
        groups.today.push(item);
      } else if (itemDate.isSame(yesterdayStart, 'day')) {
        groups.yesterday.push(item);
      } else if (itemDate.isAfter(last7DaysStart) && itemDate.isBefore(yesterdayStart)) {
        groups.last7Days.push(item);
      } else {
        groups.earlier.push(item);
      }

      return groups;
    },
    { today: [], yesterday: [], last7Days: [], earlier: [] },
  );
}

export default function ChatHistoryDrawer({ onHistoryDrawerClose }: IChatHistoryDrawerProps) {
  const { historyLoading, fetchChatHistory, chatHistory, delHistoryLoading, deleteChatHistory, setShowDeleteId, showDeleteId } = useChatHistoryDrawer();

  useEffect(() => {
    fetchChatHistory();
  }, []);

  const grouped = useMemo(() => groupChatHistoryByDate(chatHistory), [chatHistory]);

  const renderGroup = (title: string, list: any[]) => {
    if (!list.length) return null;
    return (
      <>
        <div className={styles.groupTitle}>{title}</div>
        <div className={styles.groupList}>
          {list.map((item) => (
            <div
              className={styles.historyCard}
              key={item.id}
            >
              <div className={styles.groupLeft}>
                <div className={styles.title}>{item.title}</div>
                <Tooltip title={item.modelName}>
                  <div className={styles.modelName}>{item.modelName}</div>
                </Tooltip>
              </div>
              <div className={styles.groupRight}>
                <Popconfirm
                  title={'删除后将无法查看该会话记录，确认删除吗？'}
                  destroyOnHidden={true}
                  onConfirm={() => deleteChatHistory(item.id)}
                  onCancel={() => setShowDeleteId(null)}
                  okButtonProps={{ loading: delHistoryLoading }}
                  open={showDeleteId === item.id}
                >
                  <TrashIcon
                    size={16}
                    className={styles.deleteBtn}
                    onClick={() => {
                      if (delHistoryLoading) return;
                      setShowDeleteId(item.id);
                    }}
                  />
                </Popconfirm>
              </div>
            </div>
          ))}
        </div>
      </>
    );
  };

  return (
    <Drawer
      title="历史对话"
      placement="right"
      closable={false}
      maskClosable={true}
      destroyOnHidden={true}
      onClose={onHistoryDrawerClose}
      open={true}
      loading={historyLoading}
    >
      <div className={styles.chatHistory}>
        {renderGroup('今天', grouped.today)}
        {renderGroup('昨天', grouped.yesterday)}
        {renderGroup('近7天', grouped.last7Days)}
        {renderGroup('更早', grouped.earlier)}
      </div>
    </Drawer>
  );
}
