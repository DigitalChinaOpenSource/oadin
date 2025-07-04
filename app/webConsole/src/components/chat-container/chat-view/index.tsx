import React, { useEffect, useState } from 'react';
import { ChatInput, MessageList, ChatMessageList, type MessageContentType, type ChatMessageItem, registerMessageContents } from '@res-utiles/ui-components';
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
import useModelPathChangeStore from '@/store/useModelPathChangeStore';
import useSelectedModelStore from '@/store/useSelectedModel';
import { fetchCheckEngineStatus, chechIsModelDownloaded } from '../utils';
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
  isDownloadEmbed: boolean; // 是否下载词嵌入模型
}

export default function ChatView(props: IChatViewProps) {
  const { isDownloadEmbed } = props;
  const { messages, isUploading } = useChatStore();
  const { selectedModel } = useSelectedModelStore();
  const migratingStatus = useModelPathChangeStore.getState().migratingStatus;
  const { containerRef, handleScroll, getIsNearBottom, scrollToBottom } = useScrollToBottom<HTMLDivElement>();
  const { sendChatMessage, streamingContent, streamingThinking, isLoading, isResending, error, cancelRequest, resendLastMessage } = useChatStream();

  useEffect(() => {
    // 如果消息列表有更新且当前滚动位置接近底部，则自动滚动到底部
    if (messages.length > 0 && getIsNearBottom()) {
      scrollToBottom();
    }
  }, [messages.length, getIsNearBottom, scrollToBottom]);

  // 正在生成的消息控制滚动
  const chattingMessageControlScroll = () => {
    if (getIsNearBottom()) {
      scrollToBottom();
    }
  };

  const handleSendMessage = async (messageString: string) => {
    if (!messageString.trim() || isLoading || isUploading) return;
    const isEngineAvailable = await fetchCheckEngineStatus();
    if (!isEngineAvailable) {
      message.error('模型引擎异常，请检查当前模型引擎的服务状态');
      return;
    }
    const isModelDownloaded = await chechIsModelDownloaded(selectedModel?.name || '');
    if (!isModelDownloaded) {
      const text = selectedModel?.source === 'local' ? '模型未下载，请先下载模型' : '模型未授权，请先授权';
      message.error(text);
      return;
    }
    await sendChatMessage(messageString);
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
  console.log('messages===>', messages);
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
                return (
                  <div className={'user-question'}>
                    {/*{!!attachmentFiles && (*/}
                    {/*  <div className={'user-question-files'}>*/}
                    {/*    <span>引用文件:</span>*/}
                    {/*    {(attachmentFiles as UploadFile[]).map((file) => (*/}
                    {/*      <div*/}
                    {/*        key={file.uid}*/}
                    {/*        className={'upload-file-item'}*/}
                    {/*      >*/}
                    {/*        {file.name}*/}
                    {/*      </div>*/}
                    {/*    ))}*/}
                    {/*  </div>*/}
                    {/*)}*/}
                    <MarkdownContent dataSource={content as string} />
                  </div>
                );
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

registerMessageContents({
  // @ts-ignore
  plain: MarkdownContent,
  think: (props: any) => <DeepThinkChat dataSource={props.dataSource} />,
  mcp: (props: any) => <McpToolChat dataSource={props.dataSource} />,
});
