import { PrismLight, SyntaxHighlighterProps } from 'react-syntax-highlighter';
import json from 'react-syntax-highlighter/dist/esm/languages/prism/json';
import React, { useState } from 'react';
import styles from './index.module.scss';
import { coy } from 'react-syntax-highlighter/dist/cjs/styles/prism';
import CopyIcon from '@/components/icons/copy.tsx';
import CopySuccess from '@/components/icons/copy-success.tsx';
import { message } from 'antd';
// 注册需要的语言（性能优化）
PrismLight.registerLanguage('json', json);
const SyntaxHighlighter = PrismLight as unknown as React.ComponentType<SyntaxHighlighterProps>;

export interface IJsonParsePanelProps {
  code?: string;
  maxHeight?: string;
  isConfig?: boolean; // 是否是服务器配置
  propsContentStyles?: React.CSSProperties;
}
export default function JsonParsePanel(props: IJsonParsePanelProps) {
  const { code, maxHeight, isConfig = true, propsContentStyles } = props;
  const [showDefaultIcon, setShowDefaultIcon] = useState<boolean>(true);

  const handleClick = async () => {
    try {
      if (!showDefaultIcon) return;
      message.success('复制成功');
      setShowDefaultIcon(false);
      await navigator.clipboard.writeText(code || defaultCode);
    } catch (err) {
      console.error('复制失败:', err);
    } finally {
      setTimeout(() => {
        setShowDefaultIcon(true);
      }, 3000);
    }
  };

  const defaultCode = `{
    "model": { 
      "name": "测试",
      "version": "1.0.0",
    }
  }`;

  return (
    <div className={styles.jsonParsePanel}>
      {isConfig && <div className={styles.configHeader}>服务器配置</div>}
      <div
        className={styles.configContent}
        style={propsContentStyles}
      >
        <div
          className={styles.copyIcon}
          onClick={handleClick}
        >
          {showDefaultIcon ? (
            <CopyIcon
              fill={'#C1C6D6'}
              hoverFill={'blue'}
            />
          ) : (
            <CopySuccess />
          )}
        </div>
        <SyntaxHighlighter
          language="json"
          style={coy}
          customStyle={{
            // width: '400px',
            maxHeight: maxHeight || '300px',
            fontSize: '12px',
            backgroundColor: 'transparent',
            border: 'none',
          }}
        >
          {code || defaultCode}
        </SyntaxHighlighter>
      </div>
    </div>
  );
}
