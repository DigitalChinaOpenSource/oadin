/** 深度思考对话块 */
import React, { useState } from 'react';
import styles from './index.module.scss';
import deepThinkSvg from '@/components/icons/deep-think.svg';
import { CheckCircleIcon } from '@phosphor-icons/react';
import arrowUp from '@/components/icons/arrow-up.svg';
import arrowDown from '@/components/icons/arrow-down.svg';

// interface IDeepThinkChatProps {}

export default function DeepThinkChat() {
  // TODO 思考完毕之后自动收起
  const [isExpanded, setIsExpanded] = useState(false);
  return (
    <div className={styles.deepThinkChat}>
      <div className={styles.header}>
        <div className={styles.chatStatus}>
          {/* <>
            <img
              src={deepThinkSvg}
              alt="深度思考中"
            />
            <div
              className={styles.thinkingText}
            >
                深度思考中...
            </div>
          </> */}
          {/* <>
            <img
              src={deepThinkSvg}
              alt="思考已停止"
            />
            <div className={styles.statusText}>思考已停止</div>
          </> */}
          <>
            <CheckCircleIcon
              width={16}
              height={16}
              fill="#4f4dff"
            />
            <div className={styles.statusText}>已深度思考</div>
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
        <div className={styles.resultText}>
          嗯，用户再次要求规划北京到天津的3天行程，看来对之前的方案有补充需求。可能有两种情况：一是没看到第一次回复的内容，二是希望获得更差异化的方案。考虑到历史记录里已经给出过经典路线，这次应该侧重提供不同主题的备选。
          用户可能是自由行的年轻人或家庭游客。从简洁的提问方式看，ta更关注实用信息而非背景介绍，但特意重复提问暗示对原有方案某些部分不满意——或许觉得滨海新区行程太赶？或是想避开人多的网红点？
          嗯，用户再次要求规划北京到天津的3天行程，看来对之前的方案有补充需求。可能有两种情况：一是没看到第一次回复的内容，二是希望获得更差异化的方案。考虑到历史记录里已经给出过经典路线，这次应该侧重提供不同主题的备选。
          用户可能是自由行的年轻人或家庭游客。从简洁的提问方式看，ta更关注实用信息而非背景介绍，但特意重复提问暗示对原有方案某些部分不满意——或许觉得滨海新区行程太赶？或是想避开人多的网红点？
          嗯，用户再次要求规划北京到天津的3天行程，看来对之前的方案有补充需求。可能有两种情况：一是没看到第一次回复的内容，二是希望获得更差异化的方案。考虑到历史记录里已经给出过经典路线，这次应该侧重提供不同主题的备选。
          用户可能是自由行的年轻人或家庭游客。从简洁的提问方式看，ta更关注实用信息而非背景介绍，但特意重复提问暗示对原有方案某些部分不满意——或许觉得滨海新区行程太赶？或是想避开人多的网红点？
        </div>
      </div>
    </div>
  );
}
