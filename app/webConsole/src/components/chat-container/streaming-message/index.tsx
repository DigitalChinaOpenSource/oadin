import React, { useEffect } from 'react';
import ReactMarkdown from 'react-markdown';
import DeepThinkChat from '../chat-components/deep-think-chat';
import remarkGfm from 'remark-gfm';
import rehypeRaw from 'rehype-raw';

interface StreamingMessageProps {
  content: string;
  scroll: () => void;
  thinkingContent?: string;
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
      {thinkingContent && <DeepThinkChat streamContent={thinkingContent} />}

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
