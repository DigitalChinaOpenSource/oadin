import { useRef, useState, useLayoutEffect } from 'react';
import { Tag, Tooltip } from 'antd';
import styles from './index.module.scss';

interface TagsRenderProps {
  tags: string[];
  className?: string;
}

export default function TagsRender({ tags, className }: TagsRenderProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [visibleTags, setVisibleTags] = useState<string[]>([]);
  const [hiddenTags, setHiddenTags] = useState<string[]>([]);

  useLayoutEffect(() => {
    const container = containerRef.current;
    if (!container) return;
    // 重置标签状态
    setVisibleTags([]);
    setHiddenTags([]);
    // 创建一个临时的标签元素来测量宽度
    const measureTag = (text: string): number => {
      const tempTag = document.createElement('span');
      tempTag.className = styles.tag;
      tempTag.style.position = 'absolute';
      tempTag.style.visibility = 'hidden';
      tempTag.style.display = 'inline-block';
      tempTag.innerText = text;
      document.body.appendChild(tempTag);
      const width = tempTag.offsetWidth;
      document.body.removeChild(tempTag);
      return width;
    };

    const calculateVisibleTags = () => {
      // 计算 "+n" 标签的宽度
      const countTagWidth = measureTag('+99');
      // 计算可用宽度
      const availableWidth = container.offsetWidth - countTagWidth;
      let totalWidth = 0;
      const visible: string[] = [];
      const hidden: string[] = [];
      // 计算每个标签的宽度并决定是否显示
      for (let i = 0; i < tags.length; i++) {
        const tagWidth = measureTag(tags[i]);
        if (totalWidth + tagWidth <= availableWidth) {
          totalWidth += tagWidth;
          visible.push(tags[i]);
        } else {
          hidden.push(tags[i]);
        }
      }

      setVisibleTags(visible);
      setHiddenTags(hidden);
    };
    // 初始计算
    calculateVisibleTags();
    // 监听容器大小变化
    const resizeObserver = new ResizeObserver(() => {
      calculateVisibleTags();
    });

    resizeObserver.observe(container);

    // 清理函数
    return () => {
      resizeObserver.disconnect();
    };
  }, [tags]);

  // 渲染隐藏的标签
  const hiddenTagsRender = () => {
    if (hiddenTags.length === 0) return null;
    return (
      <>
        {hiddenTags.map((tag, index) => (
          <div
            key={index}
            className={styles.hiddenTag}
          >
            {tag}
          </div>
        ))}
      </>
    );
  };
  return (
    <div
      className={`${styles.tagWrapper} ${className || ''}`}
      style={{ width: '100%' }}
      ref={containerRef}
    >
      {visibleTags.map((tag, index) => (
        <Tag
          key={index}
          className={styles.tag}
          style={{
            marginRight: 8,
            flexShrink: 0,
          }}
        >
          {tag}
        </Tag>
      ))}
      {hiddenTags.length > 0 && (
        <Tooltip title={hiddenTagsRender()}>
          <Tag
            style={{ flexShrink: 0 }}
            className={styles.tag}
          >
            +{hiddenTags.length}
          </Tag>
        </Tooltip>
      )}
    </div>
  );
}
