import { Button, Modal } from 'antd';
import remarkGfm from 'remark-gfm';
import rehypeRaw from 'rehype-raw';
import ReactMarkdown from 'react-markdown';
import styles from './index.module.scss';
interface ICopyrightNoticeProps {
  open: boolean;
  onClose: () => void;
  notice?: string;
}

export default function CopyrightNotice({ open, onClose, notice }: ICopyrightNoticeProps) {
  return (
    <Modal
      centered={true}
      closable={true}
      title={'版权声明'}
      onCancel={onClose}
      open={open}
      onOk={onClose}
      styles={{ body: { margin: '24px 0' } }}
      footer={<Button onClick={onClose}>关闭</Button>}
    >
      <div className={styles.markdownBody}>
        <>
          {ReactMarkdown({
            children: notice || '这里是版权声明内容',
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
    </Modal>
  );
}
