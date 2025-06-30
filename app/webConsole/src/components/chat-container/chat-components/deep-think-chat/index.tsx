/** 深度思考对话块 */
import React, { useState } from 'react';
import styles from './index.module.scss';
import deepThinkSvg from '@/components/icons/deep-think.svg';
import { CheckCircleIcon } from '@phosphor-icons/react';
import arrowUp from '@/components/icons/arrow-up.svg';
import arrowDown from '@/components/icons/arrow-down.svg';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeRaw from 'rehype-raw';

interface IDeepThinkChatProps {
  dataSource?: {
    data: string;
    status?: 'success' | 'error' | 'progress';
    coastTime?: number;
  };
}

export default function DeepThinkChat(props: IDeepThinkChatProps) {
  const { dataSource } = props;
  // TODO 思考完毕之后自动收起
  const [isExpanded, setIsExpanded] = useState(true);
  return (
    <div className={styles.deepThinkChat}>
      <div className={styles.header}>
        <div className={styles.chatStatus}>
          {dataSource?.status === 'progress' && (
            <>
              <img
                src={deepThinkSvg}
                alt="深度思考中"
              />
              <div className={styles.thinkingText}>深度思考中...</div>
            </>
          )}
          {dataSource?.status === 'error' && (
            <>
              <img
                src={deepThinkSvg}
                alt="思考已停止"
              />
              <div className={styles.statusText}>思考已停止</div>
            </>
          )}
          {dataSource?.status === 'success' && (
            <>
              <CheckCircleIcon
                width={16}
                height={16}
                fill="#4f4dff"
              />
              <div className={styles.statusText}>已深度思考</div>
              {dataSource?.coastTime && <div className={styles.coastTime}>（用时 {dataSource?.coastTime} 秒）</div>}
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
        <div className={styles.resultText}>
          {ReactMarkdown({
            children: dataSource?.data || '',
            remarkPlugins: [remarkGfm],
            rehypePlugins: [rehypeRaw],
          })}
        </div>
      </div>
    </div>
  );
}
