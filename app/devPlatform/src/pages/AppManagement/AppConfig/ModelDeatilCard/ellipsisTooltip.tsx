// 通用溢出时才显示Tooltip的组件
import { FC, useEffect, useRef, useState } from 'react';
import { Tooltip } from 'antd';

export default function EllipsisTooltip({
  title,
  children,
  className,
  style,
  maxWidth,
}: {
  title: string;
  children: React.ReactNode;
  className?: string;
  style?: React.CSSProperties;
  maxWidth?: number | string;
}) {
  const [isOverflow, setIsOverflow] = useState(false);
  return isOverflow ? (
    <Tooltip
      title={title}
      styles={{ root: { maxWidth } }}
    >
      <div style={{ display: 'grid' }}>
        <Content
          className={className}
          onOverflowChange={setIsOverflow}
        >
          {children}
        </Content>
      </div>
    </Tooltip>
  ) : (
    <Content
      className={className}
      onOverflowChange={setIsOverflow}
    >
      {children}
    </Content>
  );
}

const Content: FC<{
  className?: string;
  children: React.ReactNode;
  onOverflowChange?: (overflow: boolean) => void;
}> = ({ className, children, onOverflowChange }) => {
  const ref = useRef<HTMLDivElement>(null);
  const checkOverflow = (el: HTMLDivElement) => {
    // const el = ref.current;
    if (!el) return;
    const overflowing = el.scrollWidth > el.clientWidth || el.scrollHeight > el.clientHeight;
    onOverflowChange?.(overflowing);
  };

  useEffect(() => {
    if (!ref.current) return;
    // checkOverflow(); // 初始渲染后主动检测一次

    const ob = new ResizeObserver(() => checkOverflow(ref.current!)); // 创建ResizeObserver实例
    ob.observe(ref.current);
    return () => {
      ob.disconnect();
    };
  }, []); // 依赖内容变化
  return (
    <div
      ref={ref}
      className={className}
    >
      {children}
    </div>
  );
};
