import React, { useEffect, useState } from 'react';
import { ChatInput, MessageList, type MessageType, type MessageContentType, registerMessageContents } from '@res-utiles/ui-components';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeRaw from 'rehype-raw';
import '@res-utiles/ui-components/dist/index.css';
import { Button, message } from 'antd';
import { SelectMcp } from '@/components/select-mcp';
import { CopyIcon, ArrowClockwiseIcon, StopIcon } from '@phosphor-icons/react';
import DeepThinkChat from '../chat-components/deep-think-chat';
import McpToolChat from '../chat-components/mcp-tool-chat';
import StreamingMessage from '../streaming-message';
import UploadTool from '../upload-tool';
import useChatStore from '../store/useChatStore';
import sendSvg from '@/components/icons/send.svg';
import realLoading from '@/components/icons/real-loading.svg';
import { useScrollToBottom } from '@/hooks/useScrollToBottom';
import { useChatStream } from '@/components/chat-container/useChatStream';
import { copyMessageToClipboard } from '../useChatStream/utils';
import { HeaderContent } from './header-content';
import EmbedDownloadButton from '../enbed-download-btn';
import './index.scss';

interface IChatViewProps {
  isUploadVisible?: boolean; // 上传功能是否可用，是否下载词嵌入模型
}

interface ChatMessageContent extends MessageContentType {
  /**
   * 内容类型
   */
  type: 'message' | 'task' | 'canvas' | 'file' | 'ref';
}

interface ChatMessage extends MessageType {
  /**
   * 所属对话ID
   */
  conversationId: string;
  /**
   * 消息时间
   */
  createdAt?: number;
  /**
   * 消息内容列表
   */
  contentList?: ChatMessageContent[];
}

export default function ChatView({ isUploadVisible }: IChatViewProps) {
  const { messages, isUploading } = useChatStore();
  const { containerRef, handleScroll, getIsNearBottom, scrollToBottom } = useScrollToBottom<HTMLDivElement>();

  const { sendChatMessage, streamingContent, streamingThinking, isLoading, isResending, error, cancelRequest, resendLastMessage } = useChatStream();

  useEffect(() => {
    // 如果消息列表有更新且当前滚动位置接近底部，则自动滚动到底部
    if (messages.length > 0 && getIsNearBottom()) {
      scrollToBottom();
    }
  }, [messages.length, getIsNearBottom, scrollToBottom]);

  useEffect(() => {
    if (messages.length === 0 && streamingContent) {
      cancelRequest();
    }
  }, [messages.length, streamingContent, cancelRequest]);

  // 正在生成的消息控制滚动
  const chattingMessageControlScroll = () => {
    if (getIsNearBottom()) {
      scrollToBottom();
    }
  };

  const handleSendMessage = (message: string) => {
    if (!message.trim() || isLoading || isUploading) return;

    sendChatMessage(message);
  };
  // 复制消息
  const handleCopyMessage = (content?: string) => {
    if (copyMessageToClipboard(content)) {
      message.success('已复制到剪贴板');
    } else {
      message.error('复制失败');
    }
  };

  // 输入框底部功能区
  const footerContent = (
    <div style={{ display: 'flex', alignItems: 'center', gap: 16, fontSize: 14, color: '#7553FC' }}>
      {isUploadVisible ? <UploadTool /> : <EmbedDownloadButton />}

      <SelectMcp />
    </div>
  );

  // 流式消息和控制按钮的底部面板
  const renderBottomPanel = () => {
    return (
      <div className="bottom-panel-container">
        {/* 流式生成的消息 */}
        {isLoading && messages.length > 0 && (
          <StreamingMessage
            content={streamingContent}
            thinkingContent={streamingThinking}
            scroll={chattingMessageControlScroll}
          />
        )}

        {(isLoading || streamingContent) && (
          <div className="message-control-buttons">
            {isLoading && (
              <img
                src={realLoading}
                alt="加载中"
                style={{ width: 24, height: 24 }}
              />
            )}
          </div>
        )}
        {error && (
          <Button
            type="link"
            icon={<ArrowClockwiseIcon width={16} />}
            onClick={resendLastMessage}
            loading={isResending}
            disabled={isLoading && !isResending}
          >
            重试
          </Button>
        )}
        {!isLoading && messages.length > 0 && (
          <div className="message-control-buttons">
            <Button
              type="link"
              icon={<CopyIcon width={16} />}
              onClick={() => handleCopyMessage()}
            >
              复制
            </Button>
            <Button
              type="link"
              icon={<ArrowClockwiseIcon width={16} />}
              onClick={resendLastMessage}
              loading={isResending}
              disabled={isResending}
            >
              重新发送
            </Button>
          </div>
        )}
      </div>
    );
  };

  return (
    <div className="chat-layout">
      <div className="chat-body">
        <div
          className="chat-message-area chat-width"
          ref={containerRef}
          onScroll={handleScroll}
        >
          <MessageList
            messages={messages}
            setBubbleProps={(message) => ({
              align: message.role === 'user' ? 'right' : 'left',
              classNames: {
                content: message.role === 'user' ? 'user-bubble' : 'ai-bubble',
              },
            })}
            className="chat-message-list"
            contentListClassName="chat-message-content-list"
            bottomPanel={renderBottomPanel()}
          />

          {/* 错误消息展示 */}
          {/* {error && !isLoading && !isResending && (
            <div
              className="error-message"
              style={{ color: '#e85951', padding: '8px 16px', margin: '8px 0' }}
            >
              {error}
            </div>
          )} */}
        </div>

        <div className="chat-input-area chat-width">
          <ChatInput
            placeholder="输入消息..."
            onSend={handleSendMessage}
            className="chat-input"
            renderSendButton={({ onClick, inputValue }) => (
              <>
                {isLoading || isUploading ? (
                  <Button
                    icon={
                      <StopIcon
                        width={24}
                        weight="fill"
                        fill="#4f4dff"
                      />
                    }
                    onClick={() => {
                      console.log('stop');
                      cancelRequest();
                    }}
                  />
                ) : (
                  <Button
                    type="primary"
                    style={{ borderRadius: 8, cursor: 'pointer' }}
                    icon={
                      <img
                        src={sendSvg}
                        alt="发送"
                      />
                    }
                    onClick={onClick}
                  />
                )}
              </>
            )}
            header={<HeaderContent />}
            footer={footerContent}
          />
        </div>
      </div>
    </div>
  );
}

const MarkdownContent = ({ dataSource }: { dataSource?: string }) => {
  // 如果内容不是文本，或者是空白文本，直接返回空
  if (!dataSource || typeof dataSource !== 'string' || !dataSource.trim()) {
    return null;
  }

  return (
    <div className="markdown-content">
      <>
        {ReactMarkdown({
          children: dataSource,
          remarkPlugins: [remarkGfm],
          rehypePlugins: [rehypeRaw],
        })}
      </>
    </div>
  );
};

registerMessageContents({
  // @ts-ignore
  plain: MarkdownContent,
  think: (props: any) => <DeepThinkChat dataSource={props.dataSource} />,
  mcp: (props: any) => <McpToolChat dataSource={props.dataSource} />,
});
