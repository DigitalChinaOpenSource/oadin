import { PrismLight, SyntaxHighlighterProps } from 'react-syntax-highlighter';
import json from 'react-syntax-highlighter/dist/esm/languages/prism/json';
import React from 'react';
import styles from './index.module.scss';
import { coy } from 'react-syntax-highlighter/dist/cjs/styles/prism';
import copyIcon from '@/components/icons/copy.svg';
// 注册需要的语言（性能优化）
PrismLight.registerLanguage('json', json);
const SyntaxHighlighter = PrismLight as unknown as React.ComponentType<SyntaxHighlighterProps>;
interface McpServiceConfigProps {
  code?: string;
}

export default function McpServiceConfig({ code = '' }: McpServiceConfigProps) {
  const handleClick = async () => {
    try {
      await navigator.clipboard.writeText(code || defaultCode);
    } catch (err) {
      console.error('复制失败:', err);
    }
  };

  const defaultCode = `{
    "model": { 
      "name": "mcp",
      "version": "1.0.0,2sadasdaggggsfsddgfhgjhkhgggggggaadsa,afsd",
      "version": "1.0.0,2sadasdaggggsfsddgfhgjhkhgggggggaadsa,afsd",
      "version": "1.0.0,2sadasdaggggsfsddgfhgjhkhgggggggaadsa,afsd",
      "version": "1.0.0,2sadasdaggggsfsddgfhgjhkhgggggggaadsa,afsd",
      "version": "1.0.0,2sadasdaggggsfsddgfhgjhkhgggggggaadsa,afsd",
      "version": "1.0.0,2sadasdaggggsfsddgfhgjhkhgggggggaadsa,afsd",
      "version": "1.0.0,2sadasdaggggsfsddgfhgjhkhgggggggaadsa,afsd",
      "version": "1.0.0,2sadasdaggggsfsddgfhgjhkhgggggggaadsa,afsd",
      "version": "1.0.0,2sadasdaggggsfsddgfhgjhkhgggggggaadsa,afsd",
      "version": "1.0.0,2sadasdaggggsfsddgfhgjhkhgggggggaadsa,afsd",
      "version": "1.0.0,2sadasdaggggsfsddgfhgjhkhgggggggaadsa,afsd",
      "version": "1.0.0,2sadasdaggggsfsddgfhgjhkhgggggggaadsa,afsd",
      "version": "1.0.0,2sadasdaggggsfsddgfhgjhkhgggggggaadsa,afsd",
      "version": "1.0.0,2sadasdaggggsfsddgfhgjhkhgggggggaadsa,afsd",
      "version": "1.0.0,2sadasdaggggsfsddgfhgjhkhgggggggaadsa,afsd",
      "version": "1.0.0,2sadasdaggggsfsddgfhgjhkhgggggggaadsa,afsd",
      "version": "1.0.0,2sadasdaggggsfsddgfhgjhkhgggggggaadsa,afsd",
      "version": "1.0.0,2sadasdaggggsfsddgfhgjhkhgggggggaadsa,afsd",
      "version": "1.0.0,2sadasdaggggsfsddgfhgjhkhgggggggaadsa,afsd",
      "version": "1.0.0,2sadasdaggggsfsddgfhgjhkhgggggggaadsa,afsd",
      "version": "1.0.0,2sadasdaggggsfsddgfhgjhkhgggggggaadsa,afsd",
      "version": "1.0.0,2sadasdaggggsfsddgfhgjhkhgggggggaadsa,afsd",
      "version": "1.0.0,2sadasdaggggsfsddgfhgjhkhgggggggaadsa,afsd",
    }
  }`;

  return (
    <div className={styles.mcpServiceConfig}>
      <div className={styles.configHeader}>服务器配置</div>
      <div className={styles.configContent}>
        <div
          className={styles.copyIcon}
          onClick={handleClick}
        >
          <img
            src={copyIcon}
            alt="复制"
          />
        </div>
        {/*<Button onClick={handleClick}>复制</Button>*/}
        <SyntaxHighlighter
          language="json"
          style={coy}
          customStyle={{
            // width: '400px',
            maxHeight: '300px',
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
