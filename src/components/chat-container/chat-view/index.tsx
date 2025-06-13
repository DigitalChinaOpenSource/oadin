import React, { useEffect, useRef } from 'react';
import { ChatInput, MessageList, type MessageType, type MessageContentType, registerMessageContents } from '@res-utiles/ui-components';
import '@res-utiles/ui-components/dist/index.css';
import { Button, Tooltip } from 'antd';
import type { UploadFile } from 'antd';
import { SelectMcp } from '@/components/select-mcp';
import { FolderIcon, XCircleIcon } from '@phosphor-icons/react';
import DeepThinkChat from '../chat-components/deep-think-chat';
import McpToolChat from '../chat-components/mcp-tool-chat';
import UploadTool from '../upload-tool';
import useChatStore from '../store/useChatStore';
import sendSvg from '@/components/icons/send.svg';
import uploadSvg from '@/components/icons/upload.svg';
import rollingSvg from '@/components/icons/rolling.svg';
import { useDownLoad } from '@/hooks/useDownload';
import { useScrollToBottom } from '@/hooks/useScrollToBottom';
import './index.css';

interface IChatViewProps {
  isUploadVisible?: boolean; // 上传功能是否可用，是否下载词嵌入模型
}

interface ChatMessageContent extends MessageContentType {
  /**
   * 内容类型
   * @description 目前支持 message task canvas file ref 四种类型
   * - message: 文本消息
   * - task: 任务消息
   * - canvas: 画布消息
   * - file: 文件消息
   * - ref: 引用消息
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
   * @description 一个消息可以包含多个元素，比如一个消息可以包含多个文本、多个任务等
   */
  contentList?: ChatMessageContent[];
}

export default function ChatView({ isUploadVisible }: IChatViewProps) {
  const { messages, addMessage, uploadFileList, setUploadFileList } = useChatStore();
  const { containerRef, handleScroll, getIsNearBottom, scrollToBottom } = useScrollToBottom<HTMLDivElement>();
  const { fetchDownloadStart } = useDownLoad();

  useEffect(() => {
    // 如果消息列表有更新且当前滚动位置接近底部，则自动滚动到底部
    if (messages.length > 0 && getIsNearBottom()) {
      scrollToBottom();
    }
  }, [messages.length]);

  /**
   * 正在生成的消息控制滚动
   * @description 如果正在生成的消息存在且当前滚动位置接近底部，则自动滚动到底部
   */
  const chattingMessageControlScroll = (message: ChatMessage) => {
    if (message && getIsNearBottom()) {
      scrollToBottom();
    }
  };

  const handleSendMessage = (message: string) => {
    if (!message.trim()) return;

    const userMessage: MessageType = {
      id: Date.now().toString(),
      role: 'user',
      contentList: [
        {
          id: '1',
          type: 'plain',
          content: message,
        },
      ],
    };

    addMessage(userMessage);

    // 模拟AI回复，实际项目中替换为API调用
    setTimeout(() => {
      const aiMessage: MessageType = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        contentList: [
          {
            id: '1',
            type: 'plain',
            content: `你好，我收到了你的消息: "${message}"`,
          },
        ],
      };
      addMessage(aiMessage);
    }, 1000);
  };

  const onFileListChange = (fileList: UploadFile[]) => {
    setUploadFileList(fileList);
  };

  const headerContent = (
    <div style={{ display: 'flex', alignItems: 'center', gap: 16, flexWrap: 'wrap' }}>
      {uploadFileList.map((file) => (
        <div
          key={file.uid}
          className="upload-file-item"
        >
          {/* <div className="file-icon uploading-icon">
          <img
            src={rollingSvg}
            alt=""
          />
        </div> */}
          {/* <div className="file-icon done-icon">
          <FolderIcon
            width={16}
            height={16}
            fill="#ffffff"
          />
        </div> */}
          {/* <div className="file-icon error-icon">
          <XCircleIcon
            width={16}
            height={16}
            weight="fill"
            fill="#e85951"
          />
        </div> */}
          {file.name}
          <div
            className="upload-file-remove"
            onClick={(e) => {
              e.stopPropagation();
              setUploadFileList(uploadFileList.filter((item) => item.uid !== file.uid));
            }}
          >
            <XCircleIcon
              width={16}
              height={16}
              fill="#9ca3af"
              weight="fill"
            />
          </div>
        </div>
      ))}
    </div>
  );

  const footerContent = (
    <div style={{ display: 'flex', alignItems: 'center', gap: 16, fontSize: 14, color: '#7553FC' }}>
      {/* TODO 去掉感叹号 */}
      {isUploadVisible ? (
        <UploadTool
          onFileListChange={onFileListChange}
          uploadFileList={uploadFileList}
        />
      ) : (
        <Tooltip
          arrow={false}
          title={
            <>
              该功能需先下载词嵌入模型
              <a
                onClick={() => {
                  fetchDownloadStart({
                    name: 'quentinz/bge-large-zh-v1.5:f16',
                    service_name: 'embed',
                    source: 'local',
                    service_provider_name: 'local_ollama_embed',
                    id: 'bc8ca0995fcd651',
                  } as any);
                }}
              >
                【立即下载】
              </a>
            </>
          }
        >
          <Button
            icon={
              <img
                src={uploadSvg}
                alt="上传"
              />
            }
          />
        </Tooltip>
      )}

      <SelectMcp />
    </div>
  );

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
            //  bottomPanel={<ChattingMessage scroll={chattingMessageControlScroll} />}
          />
        </div>

        <div className="chat-input-area chat-width">
          <ChatInput
            placeholder="输入消息..."
            onSend={handleSendMessage}
            className="chat-input"
            SendButtonComponent={({ onClick, inputValue }) => (
              <Button
                type="primary"
                // disabled={!inputValue.trim()}
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
            header={headerContent}
            footer={footerContent}
          />
        </div>
      </div>
    </div>
  );
}

registerMessageContents({
  plain: ({ dataSource }: { dataSource?: string }) => <div style={{ fontSize: 14 }}>{dataSource}</div>,
  think: DeepThinkChat,
  mcp: McpToolChat,
});
