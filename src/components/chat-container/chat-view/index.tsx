import React, { useEffect, useRef } from 'react';
import { ChatInput, MessageList, type MessageType, registerMessageContents } from '@res-utiles/ui-components';
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
import './index.css';

interface IChatViewProps {
  isUploadVisible?: boolean; // 上传功能是否可用，是否下载词嵌入模型
}

export default function ChatView({ isUploadVisible }: IChatViewProps) {
  const { messages, addMessage, uploadFileList, setUploadFileList } = useChatStore();

  const messageAreaRef = useRef<HTMLDivElement>(null);

  // 监听消息列表变化，自动滚动
  useEffect(() => {
    if (messageAreaRef.current && messages.length > 0) {
      const scrollElement = messageAreaRef.current;
      scrollElement.scrollTop = scrollElement.scrollHeight;
    }
  }, [messages]);

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

  const headerUploadContent = (
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
      {!isUploadVisible ? (
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
              <a>【立即下载】</a>
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
          ref={messageAreaRef}
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
            // bottomPanel={<div style={{ background: '' }}>这里可以放正在生成的消息，或者是其他什么的...</div>}
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
            header={headerUploadContent}
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
