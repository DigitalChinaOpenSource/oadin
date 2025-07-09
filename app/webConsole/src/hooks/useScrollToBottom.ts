import { useRef, useCallback } from 'react';

export function useScrollToBottom<T extends HTMLElement>() {
  const containerRef = useRef<T>(null);
  const isNearBottomRef = useRef(true);
  const scrollTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const userScrollingRef = useRef(false);
  const lastScrollTimeRef = useRef(0);

  const handleScroll = useCallback(() => {
    const container = containerRef.current;
    if (!container) return;

    const now = Date.now();
    const isNearBottom = container.scrollHeight - container.scrollTop - container.clientHeight < 100;

    isNearBottomRef.current = isNearBottom;

    // 检测用户是否在主动滚动
    if (now - lastScrollTimeRef.current < 100) {
      userScrollingRef.current = true;
      setTimeout(() => {
        userScrollingRef.current = false;
      }, 500);
    }
    lastScrollTimeRef.current = now;
  }, []);

  const scrollToBottom = useCallback(() => {
    // 用户滚动就不要自动滚动
    if (userScrollingRef.current) return;

    if (scrollTimeoutRef.current) {
      clearTimeout(scrollTimeoutRef.current);
    }

    // 使用节流，避免过于频繁的滚动
    scrollTimeoutRef.current = setTimeout(() => {
      if (containerRef.current && !userScrollingRef.current) {
        containerRef.current.scrollTo({
          top: containerRef.current.scrollHeight,
          behavior: 'auto',
        });
      }
    }, 50);
  }, []);

  const scrollToBottomSmooth = useCallback(() => {
    if (containerRef.current) {
      containerRef.current.scrollTo({
        top: containerRef.current.scrollHeight,
        behavior: 'smooth',
      });
    }
  }, []);

  const getIsNearBottom = useCallback(() => isNearBottomRef.current, []);

  return {
    containerRef,
    scrollToBottom,
    scrollToBottomSmooth,
    handleScroll,
    getIsNearBottom,
  };
}
