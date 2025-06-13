import { useEffect, useMemo, useRef, useState } from 'react';
import { Drawer, Button, Tooltip, Popconfirm, Space } from 'antd';
import { useChatHistoryDrawer } from '@/components/chat-container/chat-history-drawer/view-module';
import type { MessageType } from '@res-utiles/ui-components';
import styles from './index.module.scss';
import dayjs from 'dayjs';
import { IChatHistoryItem } from '@/components/chat-container/chat-history-drawer/types.ts';
import { TrashIcon } from '@phosphor-icons/react';
import noHistorySvg from '@/components/icons/no-history.svg';
import EllipsisTooltip from '@/components/ellipsis-tooltip';
import { CloseOutlined } from '@ant-design/icons';

export interface IChatHistoryDrawerProps {
  onHistoryDrawerClose?: () => void;
  onHistorySelect?: (historyId: string, historyMessages: MessageType[]) => void;
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

export default function ChatHistoryDrawer(props: IChatHistoryDrawerProps) {
  const { onHistorySelect, onHistoryDrawerClose } = props;
  const { historyLoading, fetchChatHistory, fetchChatHistoryDetail, chatHistory, delHistoryLoading, deleteChatHistory, setShowDeleteId, showDeleteId } = useChatHistoryDrawer();

  const grouped = useMemo(() => groupChatHistoryByDate(chatHistory), [chatHistory]);

  const handleHistoryClick = async (id: string) => {
    if (delHistoryLoading) {
      return;
    }

    try {
      // 获取详细对话内容
      const historyDetail = await fetchChatHistoryDetail(id);

      // 如果提供了选择回调，则调用
      if (onHistorySelect && historyDetail) {
        onHistorySelect(id, historyDetail);
      }
    } catch (error) {
      console.error('加载历史对话失败:', error);
      // 可以添加错误提示
    }
  };

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
              onClick={() => {
                if (delHistoryLoading) {
                  // setShowDeleteId(null);
                  return;
                }
                // fetchChatHistoryDetail(item.id);
                handleHistoryClick(item.id);
              }}
            >
              <div className={styles.groupLeft}>
                <div className={styles.title}>{item.title}</div>
                {/* 修改：仅溢出时显示Tooltip */}
                <EllipsisTooltip
                  title={item.modelName}
                  className={styles.modelName}
                >
                  {item.modelName}
                </EllipsisTooltip>
              </div>
              <div className={styles.groupRight}>
                <Popconfirm
                  title={'删除后将无法查看该会话记录，确认删除吗？'}
                  destroyOnHidden={true}
                  onConfirm={(e) => {
                    e?.stopPropagation();
                    e?.preventDefault();
                    deleteChatHistory(item.id);
                  }}
                  onCancel={(e) => {
                    e?.stopPropagation();
                    e?.preventDefault();
                    setShowDeleteId(null);
                  }}
                  okButtonProps={{ loading: delHistoryLoading }}
                  cancelButtonProps={{ disabled: delHistoryLoading }}
                  open={showDeleteId === item.id}
                >
                  <TrashIcon
                    size={16}
                    className={styles.deleteBtn}
                    onClick={(e) => {
                      e.stopPropagation();
                      e.preventDefault();
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
      maskClosable={!delHistoryLoading}
      destroyOnHidden={true}
      onClose={onHistoryDrawerClose}
      open={true}
      loading={historyLoading}
      extra={
        <CloseOutlined
          className={styles.closeIcon}
          onClick={() => {
            if (delHistoryLoading) return;
            onHistoryDrawerClose?.();
          }}
        />
      }
    >
      <div className={styles.chatHistory}>
        {chatHistory.length > 0 ? (
          <>
            {renderGroup('今天', grouped.today)}
            {renderGroup('昨天', grouped.yesterday)}
            {renderGroup('近7天', grouped.last7Days)}
            {renderGroup('更早', grouped.earlier)}
          </>
        ) : (
          <div className={styles.noData}>
            <div className={styles.noDataIcon}>
              <img
                src={noHistorySvg}
                alt="no-data"
              />
            </div>
            <div className={styles.noDataText}>暂无历史对话</div>
          </div>
        )}
      </div>
    </Drawer>
  );
}
