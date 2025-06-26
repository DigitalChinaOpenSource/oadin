import React, { useEffect } from 'react';
import ReactMarkdown from 'react-markdown';
import DeepThinkChat from '../chat-components/deep-think-chat';
import remarkGfm from 'remark-gfm';
import rehypeRaw from 'rehype-raw';

interface StreamingMessageProps {
  content: string;
  scroll: () => void;
  thinkingContent?: string | { data: string; status: string };
}

const StreamingMessage: React.FC<StreamingMessageProps> = ({ content, scroll, thinkingContent }) => {
  useEffect(() => {
    if (content) {
      scroll();
    }
  }, [content, scroll]);

  if (!content && !thinkingContent) return null;

  return (
    <div className="chat-message-bubble ai-bubble">
      <div style={{ marginBottom: '16px' }}>
        {typeof thinkingContent === 'string' ? thinkingContent : thinkingContent?.data && <DeepThinkChat dataSource={{ data: thinkingContent.data, status: thinkingContent.status as any }} />}
      </div>
      {/* @ts-ignore */}
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[rehypeRaw]}
      >
        {content || ''}
      </ReactMarkdown>
    </div>
  );
};

export default StreamingMessage;
