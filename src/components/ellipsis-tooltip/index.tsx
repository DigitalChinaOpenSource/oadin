// 通用溢出时才显示Tooltip的组件
import { useEffect, useRef, useState } from 'react';
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
  const ref = useRef<HTMLDivElement>(null);
  const [isOverflow, setIsOverflow] = useState(false);

  useEffect(() => {
    const el = ref.current;
    if (el) {
      setIsOverflow(el.scrollWidth > el.clientWidth || el.scrollHeight > el.clientHeight);
    }
  }, [title, children]);

  const content = (
    <div
      ref={ref}
      className={className}
    >
      {children}
    </div>
  );

  return isOverflow ? (
    <Tooltip
      title={title}
      styles={{ root: { maxWidth } }}
    >
      {content}
    </Tooltip>
  ) : (
    content
  );
}
