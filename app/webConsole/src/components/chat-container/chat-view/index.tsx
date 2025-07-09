import React, { useEffect, useCallback } from 'react';
import { ChatInput, ChatMessageList, type MessageContentType, type ChatMessageItem } from '@res-utiles/ui-components';
import '@res-utiles/ui-components/dist/index.css';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeRaw from 'rehype-raw';
import '@res-utiles/ui-components/dist/index.css';
import { Button, message, type UploadFile } from 'antd';
import { SelectMcp } from '@/components/select-mcp';
import { CopyIcon, ArrowClockwiseIcon, StopIcon } from '@phosphor-icons/react';
import { CopyToClipboard } from 'react-copy-to-clipboard';
import DeepThinkChat from '../chat-components/deep-think-chat';
import McpToolChat from '../chat-components/mcp-tool-chat';
import StreamingMessage from '../streaming-message';
import UploadTool from '../upload-tool';
import useChatStore from '../store/useChatStore';
import sendSvg from '@/components/icons/send.svg';
import realLoading from '@/components/icons/real-loading.svg';
import { useScrollToBottom } from '@/hooks/useScrollToBottom';
import { useChatStream } from '@/components/chat-container/useChatStream';
import { HeaderContent } from './header-content';
import EmbedDownloadButton from '../enbed-download-btn';
import { useModelPathChangeStore } from '@/store/useModelPathChangeStore';
import useSelectedModelStore from '@/store/useSelectedModel';
import { fetchCheckEngineStatus, checkIsModelDownloaded } from '../utils';
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

interface IChatViewProps {
  isDownloadEmbed: boolean;
}

export default function ChatView(props: IChatViewProps) {
  const { isDownloadEmbed } = props;
  const { messages, isUploading } = useChatStore();
  const { selectedModel } = useSelectedModelStore();
  const migratingStatus = useModelPathChangeStore.getState().migratingStatus;
  const { containerRef, handleScroll, getIsNearBottom, scrollToBottom, scrollToBottomSmooth, forceScrollToBottom } = useScrollToBottom<HTMLDivElement>();
  const { sendChatMessage, streamingContent, streamingThinking, isLoading, isResending, error, cancelRequest, resendLastMessage } = useChatStream();

  // 消息列表更新时的滚动处理
  useEffect(() => {
    if (messages.length > 0 && getIsNearBottom()) {
      scrollToBottomSmooth();
    }
  }, [messages.length, getIsNearBottom, scrollToBottomSmooth]);

  // 流式消息的滚动处理
  const chattingMessageControlScroll = useCallback(() => {
    if (getIsNearBottom()) {
      scrollToBottom();
    }
  }, [getIsNearBottom, scrollToBottom]);

  const handleSendMessage = async (messageString: string) => {
    if (!messageString.trim() || isLoading || isUploading) return;

    if (selectedModel?.source === 'local') {
      const isEngineAvailable = await fetchCheckEngineStatus();
      if (!isEngineAvailable) {
        message.error('发送失败，暂无法使用模型相关功能，请稍后再试');
        return;
      }
    }

    const isModelDownloaded = await checkIsModelDownloaded(selectedModel?.name || '');
    if (!isModelDownloaded) {
      const text = selectedModel?.source === 'local' ? '模型未下载，请先下载模型' : '模型未授权，请先授权';
      message.error(text);
      return;
    }

    await sendChatMessage(messageString);

    setTimeout(() => {
      forceScrollToBottom();
    }, 100);
  };

  // 输入框底部功能区
  const footerContent = (
    <div style={{ display: 'flex', alignItems: 'center', gap: 16, fontSize: 14, color: '#7553FC' }}>
      {isDownloadEmbed ? <UploadTool /> : <EmbedDownloadButton />}

      <SelectMcp />
    </div>
  );

  const copiedFormMessage = (messages: ChatMessageItem[]) => {
    if (!messages || messages.length === 0) return '';

    const lastMessage = messages[messages.length - 1];
    if (!lastMessage?.contentList || lastMessage.contentList.length === 0) return '';

    // 收集所有内容
    const contents: string[] = [];

    // 处理深度思考内容
    for (const item of lastMessage.contentList) {
      if (item.type === 'think' && item.content && typeof item.content === 'object' && 'data' in item.content) {
        contents.push(`${item.content.data}`);
      } else if (item.type === 'plain' && typeof item.content === 'string') {
        // 排除"回复已中断"等特殊字符
        const cleanedContent = item.content.replace(/「回复已中断」/g, '').trim();
        if (cleanedContent) {
          contents.push(cleanedContent);
        }
      }
    }
    const fullContent = contents.join('\n\n');
    return fullContent || '';
  };

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
        <div className="message-control-buttons">
          {(isLoading || streamingContent) && (
            <>
              {isLoading && (
                <img
                  src={realLoading}
                  alt="加载中"
                  style={{ width: 24, height: 24 }}
                />
              )}
            </>
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
          {!error && !isLoading && messages.length > 0 && (
            <>
              <CopyToClipboard
                text={copiedFormMessage(messages)}
                onCopy={() => {
                  message.success('已复制到剪贴板');
                }}
              >
                <Button
                  type="link"
                  icon={<CopyIcon width={16} />}
                >
                  复制
                </Button>
              </CopyToClipboard>
              <Button
                type="link"
                icon={<ArrowClockwiseIcon width={16} />}
                onClick={resendLastMessage}
                disabled={isResending}
              >
                重新发送
              </Button>
            </>
          )}
        </div>
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
          <ChatMessageList
            className="chat-message-list"
            dataSource={messages}
            bubbleProps={(message) => ({
              align: message.role === 'user' ? 'right' : 'left',
              classNames: {
                content: message.role === 'user' ? 'user-bubble' : 'ai-bubble',
              },
            })}
            renderContent={(messageContent) => {
              const { content, type, attachmentFiles } = messageContent;
              if (type === 'plain') {
                return <MarkdownContent dataSource={content as string} />;
                // return !!attachmentFiles ? (
                //   <div className={'user-question'}>
                //     <div className={'user-question-files'}>
                //       <span>引用文件:</span>
                //       {(attachmentFiles as UploadFile[]).map((file) => (
                //         <div
                //           key={file.uid}
                //           className={'upload-file-item'}
                //         >
                //           {file.name}
                //         </div>
                //       ))}
                //     </div>
                //     <MarkdownContent dataSource={content as string} />
                //   </div>
                // ) : (
                //   <MarkdownContent dataSource={content as string} />
                // );
              }
              if (type === 'think') {
                return <DeepThinkChat dataSource={content as any} />;
              }
              if (type === 'mcp') {
                return <McpToolChat dataSource={content as any} />;
              }
              return null;
            }}
          ></ChatMessageList>
          <>{renderBottomPanel()}</>
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
            textareaMaxHeight={70}
            renderSendButton={({ onClick, inputValue }) => (
              <>
                {isLoading ? (
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
                    disabled={migratingStatus === 'pending' || isUploading}
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
