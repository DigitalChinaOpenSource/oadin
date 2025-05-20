import { PrismLight, SyntaxHighlighterProps } from 'react-syntax-highlighter';
import json from 'react-syntax-highlighter/dist/esm/languages/prism/json';
import { xonokai } from 'react-syntax-highlighter/dist/esm/styles/prism';
import { Button } from 'antd';
import React from 'react';
// 注册需要的语言（性能优化）
PrismLight.registerLanguage('json', json);
const SyntaxHighlighter = PrismLight as unknown as React.ComponentType<SyntaxHighlighterProps>;
interface McpServiceConfigProps {
  code?: string;
}

export default function McpServiceConfig({ code = '' }: McpServiceConfigProps) {
  const handleClick = async () => {
    try {
      await navigator.clipboard.writeText(code);
    } catch (err) {
      console.error('复制失败:', err);
    }
  };

  const defaultCode = `{
    "model": { 
      "name": "mcp",
      "version": "1.0.0"
    }
  }`;

  return (
    <div>
      <SyntaxHighlighter
        language="json"
        style={xonokai}
        customStyle={{
          width: '400px',
          height: '400px',
          fontSize: '12px',
          backgroundColor: '#fff',
        }}
      >
        {code || defaultCode}
      </SyntaxHighlighter>
      <Button onClick={handleClick}>复制</Button>
    </div>
  );
}
