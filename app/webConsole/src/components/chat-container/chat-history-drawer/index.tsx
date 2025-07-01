import { useMemo } from 'react';
import { Drawer, Popconfirm } from 'antd';
import { useChatHistoryDrawer } from '@/components/chat-container/chat-history-drawer/view-model';
import styles from './index.module.scss';
import { TrashIcon } from '@phosphor-icons/react';
import noHistorySvg from '@/components/icons/no-history.svg';
import EllipsisTooltip from '@vanta/ellipsis-tooltip';
import { CloseOutlined } from '@ant-design/icons';
import { IChatHistoryDrawerProps } from './types';
import { getSessionIdFromUrl } from '@/utils/sessionParamUtils';

export default function ChatHistoryDrawer(props: IChatHistoryDrawerProps) {
  const { historyLoading, chatHistory, delHistoryLoading, deleteChatHistory, setShowDeleteId, showDeleteId, groupChatHistoryByDate, handleHistoryClick, onHistoryDrawerClose } =
    useChatHistoryDrawer(props);

  const grouped = useMemo(() => groupChatHistoryByDate(chatHistory), [chatHistory]);

  const renderGroup = (title: string, list: any[]) => {
    if (!list.length) return null;

    return (
      <>
        <div className={styles.groupTitle}>{title}</div>
        <div className={styles.groupList}>
          {list.map((item) => (
            <div
              className={`${styles.historyCard} ${item.id === getSessionIdFromUrl() ? styles.currentItem : ''}`}
              key={item.id}
              onClick={() => {
                if (delHistoryLoading) {
                  return;
                }
                handleHistoryClick(item.id);
              }}
            >
              <div className={styles.groupLeft}>
                {/*<div className={styles.title}>{item.title || '新对话'}</div>*/}
                <EllipsisTooltip
                  title={item.title || '新对话'}
                  className={styles.title}
                >
                  {item.title || '新对话'}
                </EllipsisTooltip>
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
