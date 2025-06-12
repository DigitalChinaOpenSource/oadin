import React, { useEffect } from 'react';

interface StreamingMessageProps {
  content: string;
  thinking?: string;
  scroll: (message: any) => void;
}

const StreamingMessage: React.FC<StreamingMessageProps> = ({ content, thinking, scroll }) => {
  useEffect(() => {
    if (content) {
      scroll(content);
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
      <div style={{ fontSize: 14 }}>{content}</div>
    </div>
  );
};

export default StreamingMessage;
