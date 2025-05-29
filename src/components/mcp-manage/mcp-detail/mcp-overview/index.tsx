import styles from './index.module.scss';
import ReactMarkdown, { Options } from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeRaw from 'rehype-raw';
type markDownDataType =
  | {
      src?: string;
      zh: string;
      en?: string;
    }
  | undefined;

export default function McpOverview({ markDownData, offHeight }: { markDownData: markDownDataType; offHeight: number }) {
  // 类型断言解决方案
  // const markdownOptions: Options = {
  //   children: markDownData?.zh || markDownData?.src || '',
  //   remarkPlugins: [remarkGfm],
  //   rehypePlugins: [rehypeRaw],
  // };

  return (
    <div
      className={styles.mcpOverview}
      style={{ height: `calc(100vh - ${offHeight + 187}px)` }}
    >
      <ReactMarkdown
        remarkPlugins={[remarkGfm]}
        rehypePlugins={[rehypeRaw]}
      >
        {markDownData?.zh || markDownData?.src}
      </ReactMarkdown>
      {/*<ReactMarkdown {...markdownOptions} />*/}
    </div>
  );
}
