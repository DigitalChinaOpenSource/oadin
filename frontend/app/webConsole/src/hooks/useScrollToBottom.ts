import { useRef, useCallback, useEffect } from 'react';

export function useScrollToBottom<T extends HTMLElement>() {
  const containerRef = useRef<T>(null);
  const isNearBottomRef = useRef(true);
  const userScrollingRef = useRef(false);
  const lastScrollTopRef = useRef(0);
  const lastScrollTimeRef = useRef(0);
  const autoScrollingRef = useRef(false);
  const scrollEndTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const handleScroll = useCallback(() => {
    const container = containerRef.current;
    if (!container) return;

    const now = Date.now();
    const currentScrollTop = container.scrollTop;
    const isNearBottom = container.scrollHeight - container.scrollTop - container.clientHeight < 100;

    // 更新底部状态
    isNearBottomRef.current = isNearBottom;

    if (autoScrollingRef.current) {
      lastScrollTopRef.current = currentScrollTop;
      lastScrollTimeRef.current = now;
      return;
    }

    // 检测用户是否在主动滚动
    const timeDiff = now - lastScrollTimeRef.current;
    const scrollDiff = Math.abs(currentScrollTop - lastScrollTopRef.current);

    // 如果滚动距离较大或者时间间隔较短，是用户主动滚动
    if (scrollDiff > 10 && timeDiff < 150) {
      userScrollingRef.current = true;

      // 清除之前的定时器
      if (scrollEndTimeoutRef.current) {
        clearTimeout(scrollEndTimeoutRef.current);
      }

      // 用户停止滚动500ms后重置状态
      scrollEndTimeoutRef.current = setTimeout(() => {
        userScrollingRef.current = false;
      }, 500);
    }

    lastScrollTopRef.current = currentScrollTop;
    lastScrollTimeRef.current = now;
  }, []);

  const scrollToBottom = useCallback(() => {
    const container = containerRef.current;
    if (!container) return;

    // 如果用户正在滚动，不要自动滚动
    if (userScrollingRef.current) return;

    // 标记为自动滚动
    autoScrollingRef.current = true;

    requestAnimationFrame(() => {
      if (container && !userScrollingRef.current) {
        container.scrollTo({
          top: container.scrollHeight,
          behavior: 'auto',
        });
      }

      // 短暂延迟后重置自动滚动标记
      setTimeout(() => {
        autoScrollingRef.current = false;
      }, 50);
    });
  }, []);

  const scrollToBottomSmooth = useCallback(() => {
    const container = containerRef.current;
    if (!container) return;

    // 平滑滚动不受用户滚动状态影响，但需要标记
    autoScrollingRef.current = true;

    container.scrollTo({
      top: container.scrollHeight,
      behavior: 'smooth',
    });

    // 平滑滚动需要更长的时间来完成
    setTimeout(() => {
      autoScrollingRef.current = false;
    }, 500);
  }, []);

  const forceScrollToBottom = useCallback(() => {
    const container = containerRef.current;
    if (!container) return;

    // 强制滚动，忽略用户滚动状态
    autoScrollingRef.current = true;
    userScrollingRef.current = false;

    requestAnimationFrame(() => {
      if (container) {
        container.scrollTo({
          top: container.scrollHeight,
          behavior: 'auto',
        });
      }

      setTimeout(() => {
        autoScrollingRef.current = false;
      }, 50);
    });
  }, []);

  const getIsNearBottom = useCallback(() => isNearBottomRef.current, []);

  // 清理定时器
  useEffect(() => {
    return () => {
      if (scrollEndTimeoutRef.current) {
        clearTimeout(scrollEndTimeoutRef.current);
      }
    };
  }, []);

  return {
    containerRef,
    scrollToBottom,
    scrollToBottomSmooth,
    forceScrollToBottom,
    handleScroll,
    getIsNearBottom,
  };
}
