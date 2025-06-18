import React, { useEffect } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeRaw from 'rehype-raw';

interface StreamingMessageProps {
  content: string;
  scroll: () => void;
}

const StreamingMessage: React.FC<StreamingMessageProps> = ({ content, scroll }) => {
  useEffect(() => {
    if (content) {
      scroll();
    }
  }, [content, scroll]);

  if (!content) return null;

  return (
    <div className="chat-message-bubble ai-bubble">
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
