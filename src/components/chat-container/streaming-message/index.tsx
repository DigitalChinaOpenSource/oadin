import React, { useEffect } from 'react';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeRaw from 'rehype-raw';

interface StreamingMessageProps {
  content: string;
  thinking?: string;
  scroll: () => void;
}

const StreamingMessage: React.FC<StreamingMessageProps> = ({ content, thinking, scroll }) => {
  useEffect(() => {
    if (content) {
      scroll();
    }
  }, [content, scroll]);

  if (!content && !thinking) return null;

  return (
    <div className="chat-message-bubble ai-bubble">
      {thinking && (
        <div
          className="thinking-content"
          style={{ color: '#6b7280', marginBottom: 8, fontStyle: 'italic' }}
        >
          <div style={{ fontWeight: 'bold', marginBottom: 4 }}>思考中...</div>
          <div style={{ fontSize: 14 }}>{thinking}</div>
        </div>
      )}

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
