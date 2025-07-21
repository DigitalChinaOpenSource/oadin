import React, { useEffect, useRef } from 'react';
import DeepThinkChat from '../chat-components/deep-think-chat';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeRaw from 'rehype-raw';

interface StreamingMessageProps {
  content: string;
  scroll: () => void;
  thinkingContent?: string | { data: string; status: string };
}

const StreamingMessage: React.FC<StreamingMessageProps> = ({ content, scroll, thinkingContent }) => {
  const lastContentLengthRef = useRef(0);

  useEffect(() => {
    // 只有当内容长度大量增加时才滚动
    if (content && content.length > lastContentLengthRef.current + 10) {
      lastContentLengthRef.current = content.length;
      scroll();
    }
  }, [content, scroll]);

  // 思考内容变化时也需要滚动
  useEffect(() => {
    if (thinkingContent) {
      scroll();
    }
  }, [thinkingContent, scroll]);

  if (!content && !thinkingContent) return null;

  return (
    <div className="chat-message-bubble ai-bubble">
      <div>{typeof thinkingContent === 'string' ? thinkingContent : thinkingContent?.data && <DeepThinkChat dataSource={{ data: thinkingContent.data, status: thinkingContent.status as any }} />}</div>
      <div
        className="markdown-content"
        style={{ marginTop: typeof thinkingContent === 'string' ? thinkingContent : thinkingContent?.data ? '16px' : 'unset' }}
      >
        <>
          {ReactMarkdown({
            children: content || '',
            remarkPlugins: [remarkGfm],
            rehypePlugins: [rehypeRaw],
            components: {
              a: ({ node, ...props }) => {
                return (
                  <a
                    {...props}
                    target="_blank"
                    rel="noopener noreferrer"
                  />
                );
              },
            },
          })}
        </>
      </div>
    </div>
  );
};

export default StreamingMessage;
