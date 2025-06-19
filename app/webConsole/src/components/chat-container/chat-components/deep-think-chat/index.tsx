/** 深度思考对话块 */
import React, { useState } from 'react';
import styles from './index.module.scss';
import deepThinkSvg from '@/components/icons/deep-think.svg';
import { CheckCircleIcon } from '@phosphor-icons/react';
import arrowUp from '@/components/icons/arrow-up.svg';
import arrowDown from '@/components/icons/arrow-down.svg';

// interface IDeepThinkChatProps {}
interface IDeepThinkChatProps {
  dataSource?: {
    data: string;
    status?: 'success' | 'error' | 'progress';
  };
  streamContent?: string;
}

export default function DeepThinkChat(props: IDeepThinkChatProps) {
  console.log('DeepThinkChat dataSource=====>', props);
  const { dataSource, streamContent } = props;
  // TODO 思考完毕之后自动收起
  const [isExpanded, setIsExpanded] = useState(true);
  return (
    <div className={styles.deepThinkChat}>
      <div className={styles.header}>
        <div className={styles.chatStatus}>
          {status === 'progress' && (
            <>
              <img
                src={deepThinkSvg}
                alt="深度思考中"
              />
              <div className={styles.thinkingText}>深度思考中...</div>
            </>
          )}
          {status === 'error' && (
            <>
              <img
                src={deepThinkSvg}
                alt="思考已停止"
              />
              <div className={styles.statusText}>思考已停止</div>
            </>
          )}
          {status === 'success' && (
            <>
              <CheckCircleIcon
                width={16}
                height={16}
                fill="#4f4dff"
              />
              <div className={styles.statusText}>已深度思考</div>
              <div className={styles.coastTime}>（用时 baba 秒）</div>
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
        <div className={styles.resultText}>{streamContent || dataSource?.data}</div>
      </div>
    </div>
  );
}
