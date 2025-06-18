import { useRef } from 'react';

export function useScrollToBottom<T extends HTMLElement>() {
  const containerRef = useRef<T>(null);
  const isNearBottomRef = useRef(true);

  const handleScroll = () => {
    const container = containerRef.current;
    if (!container) return;

    const isNearBottom = container.scrollHeight - container.scrollTop - container.clientHeight < 100;

    isNearBottomRef.current = isNearBottom;
  };

  const scrollToBottom = () => {
    requestAnimationFrame(() => {
      if (containerRef.current) {
        containerRef.current.scrollTo({
          top: containerRef.current.scrollHeight,
          behavior: 'smooth',
        });
      }
    });
  };

  const getIsNearBottom = () => isNearBottomRef.current;

  return {
    containerRef,
    scrollToBottom,
    handleScroll,
    getIsNearBottom,
  };
}
