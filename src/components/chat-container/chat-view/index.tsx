import { ChatInput, MessageList, type MessageType, type ChatInputProps, registerMessageContents } from '@res-utiles/ui-components';
import '@res-utiles/ui-components/dist/index.css';
import { Button } from 'antd';
import { SelectMcp } from '@/components/select-mcp';
import DeepThinkChat from '../chat-components/deep-think-chat';
import McpToolChat from '../chat-components/mcp-tool-chat';
import UploadTool from '../upload-tool';
import sendSvg from '@/components/icons/send.svg';
import './index.css';

const testMessages: MessageType[] = [
  {
    id: '1',
    role: 'user',
    contentList: [
      {
        id: '1',
        type: 'plain',
        content: '你好AI',
      },
    ],
  },
  {
    id: '2',
    role: 'assistant',
    contentList: [
      {
        id: '1',
        type: 'plain',
        content: '你好 我是智能体',
      },
      {
        id: '2',
        type: 'think',
        content: {
          status: 'success',
          answer: '好的，用户问“明天天气怎么样”，我需要回答这个问题。',
          duration: 25,
        },
      },
      {
        id: '3',
        type: 'think',
        content: {
          status: 'thinking',
          answer: '',
          duration: 25,
        },
      },
      {
        id: '4',
        type: 'plain',
        content: '你好 我是智能体',
      },
      {
        id: '5',
        type: 'think',
        content: {
          status: 'success',
          answer: '好的，用户问“明天天气怎么样”，我需要回答这个问题。',
          duration: 25,
        },
      },
      {
        id: '6',
        type: 'think',
        content: {
          status: 'thinking',
          answer: '',
          duration: 25,
        },
      },
      {
        id: '11',
        type: 'plain',
        content: '你好 我是智能体',
      },
      {
        id: '12',
        type: 'think',
        content: {
          status: 'success',
          answer: '好的，用户问“明天天气怎么样”，我需要回答这个问题。',
          duration: 25,
        },
      },
      {
        id: '13',
        type: 'think',
        content: {
          status: 'thinking',
          answer: '',
          duration: 25,
        },
      },
      {
        id: '14',
        type: 'plain',
        content: '你好 我是智能体',
      },
      {
        id: '15',
        type: 'mcp',
        content: {
          status: 'success',
          answer: '好的，用户问“明天天气怎么样”，我需要回答这个问题。',
          duration: 25,
        },
      },
      {
        id: '16',
        type: 'mcp',
        content: {
          status: 'thinking',
          answer: '',
          duration: 25,
        },
      },
    ],
  },
  {
    id: '3',
    role: 'user',
    contentList: [
      {
        id: '1',
        type: 'plain',
        content: '你好AI，我想知道明天的天气如何？',
      },
    ],
  },
];
export default function ChatView() {
  return (
    <div className="chat-layout">
      <div className="chat-body">
        <div className="chat-message-area chat-width">
          <MessageList
            messages={testMessages}
            setBubbleProps={(message) => ({
              align: message.role === 'user' ? 'right' : 'left',
              classNames: {
                content: message.role === 'user' ? 'user-bubble' : 'ai-bubble',
              },
            })}
            contentListUlProps={{
              style: { display: 'flex', flexDirection: 'column', gap: 16 },
            }}
          />
        </div>
        <div className="chat-input-area chat-width">
          <ChatInput
            placeholder="输入消息..."
            onSend={(message) => {
              console.log('send message2222', message);
            }}
            containerProps={{
              color: '#FFF',
              borderWidth: 0,
              padding: '16px',
              style: {
                height: '100%',
              },
              borderRadius: 8,
            }}
            // disabled
            SendButtonComponent={({ onClick }) => (
              <Button
                type="primary"
                style={{
                  borderRadius: 8,
                  cursor: 'pointer',
                }}
                // TODO
                // disabled={false}
                icon={
                  <img
                    src={sendSvg}
                    alt="发送"
                  />
                }
                onClick={onClick}
              />
            )}
            // 输入框顶部扩展
            header={
              <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                <div
                  style={{
                    border: '1px solid #E4E6EB',
                    borderRadius: 8,
                    padding: '5px 12px',
                  }}
                >
                  XXX.pdf
                </div>
                <div
                  style={{
                    border: '1px solid #E4E6EB',
                    borderRadius: 8,
                    padding: '5px 12px',
                  }}
                >
                  LLL.pdf
                </div>
              </div>
            }
            // 输入框底部扩展
            extra={
              <div
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 16,
                  fontSize: 14,
                  color: '#7553FC',
                }}
              >
                <UploadTool />
                <SelectMcp />
              </div>
            }
          />
        </div>
      </div>
    </div>
  );
}

const TextContent = ({ dataSource }: { dataSource: string }) => {
  return <div style={{ fontSize: 14 }}>{dataSource}</div>;
};

/** 深度思考内容组件（示例） */
const ThinkContent = (props: { dataSource: { status: string; answer: string; duration: number } }) => {
  const { status, answer, duration } = props.dataSource;
  const renderStatus = () => {
    if (status === 'success') {
      return <div>✅已完成深度思考{`用时${duration}s`}</div>;
    }
    if (status === 'error') {
      return <div>❌深度思考出错</div>;
    }
    if (status === 'thinking') {
      return <div>⏳深度思考中</div>;
    }
    return <div>⏳深度思考中</div>;
  };
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        gap: 4,
        padding: 16,
        background: '#FFFFFF',
        color: '#898EA3',
        fontSize: 14,
      }}
    >
      {renderStatus()}
      <div>{answer}</div>
    </div>
  );
};

// 注册消息内容组件
registerMessageContents({
  // 纯文本
  plain: TextContent,
  // 深度思考
  think: DeepThinkChat,
  // mcp
  mcp: McpToolChat,
});
