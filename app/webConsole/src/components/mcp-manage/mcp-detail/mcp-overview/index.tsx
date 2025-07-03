import styles from './index.module.scss';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeRaw from 'rehype-raw';
type markDownDataType =
  | {
      src?: string;
      zh: string;
      en?: string;
    }
  | undefined;

export default function McpOverview({ markDownData }: { markDownData: markDownDataType }) {
  return (
    <div className={styles.mcpOverview}>
      <>
        {ReactMarkdown({
          children: markDownData?.zh || markDownData?.src || '',
          remarkPlugins: [remarkGfm],
          rehypePlugins: [rehypeRaw],
          components: {
            // 自定义链接渲染逻辑
            a: ({ node, ...props }) => {
              // 确保所有链接都在新窗口打开
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
  );
}
