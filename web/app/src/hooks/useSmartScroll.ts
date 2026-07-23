import { useCallback, useEffect, useRef, useState } from 'react';

export interface UseSmartScrollOptions {
  /**
   * 容器的选择器（支持 CSS 选择器字符串或直接传入 HTMLElement）
   * @default '.conversation-container'
   */
  container?: string | HTMLElement | (() => HTMLElement | null);

  /**
   * 距离底部的阈值（像素），在此范围内认为用户在底部
   * @default 10
   */
  threshold?: number;

  /**
   * 滚动行为
   * @default 'smooth'
   */
  behavior?: ScrollBehavior;

  /**
   * 是否启用智能滚动
   * @default true
   */
  enabled?: boolean;

  /**
   * 用户交互后恢复自动滚动的防抖时间（毫秒）
   * @default 150
   */
  resumeDebounceMs?: number;
}

export interface UseSmartScrollReturn {
  /**
   * 当前是否应该自动滚动
   */
  shouldAutoScroll: boolean;

  /**
   * 滚动到底部（会根据 shouldAutoScroll 判断是否执行）
   */
  scrollToBottom: () => void;

  /**
   * 强制滚动到底部（忽略 shouldAutoScroll 状态）
   */
  forceScrollToBottom: () => void;

  /**
   * 手动设置是否应该自动滚动
   */
  setShouldAutoScroll: (value: boolean) => void;

  /**
   * 检查当前是否在底部
   */
  checkIfAtBottom: () => boolean;

  /**
   * 获取容器元素
   */
  getContainer: () => HTMLElement | null;
}

/**
 * 智能滚动 Hook
 */
export function useSmartScroll(
  options: UseSmartScrollOptions = {},
): UseSmartScrollReturn {
  const {
    container = '.conversation-container',
    threshold = 10,
    behavior = 'smooth',
    enabled = true,
    resumeDebounceMs = 150,
  } = options;

  const [shouldAutoScroll, setShouldAutoScroll] = useState(true);
  /**
   * 使用 ref 存储同步的自动滚动标志，避免异步状态更新导致的竞争条件
   *
   * 场景说明：
   * 1. SSE 流式输出内容，触发 scrollToBottom()
   * 2. 用户向上滚动，触发用户交互事件
   * 3. 交互事件调用 setShouldAutoScroll(false) - 这是异步的
   * 4. 但在状态更新前，又有新的 SSE 内容到达，再次触发 scrollToBottom()
   * 5. 此时 shouldAutoScroll 状态可能还是 true，导致意外滚动
   *
   * 解决方案：
   * - ref 的更新是同步的，用户交互事件会立即更新 ref
   * - scrollToBottom() 检查 ref 而不是 state，确保获取最新值
   * - state 仍然保留，用于可能需要响应式更新的场景
   */
  const shouldAutoScrollRef = useRef(true);
  const containerRef = useRef<HTMLElement | null>(null);
  const userInteractingRef = useRef(false); // 标记用户是否正在交互
  const resumeTimerRef = useRef<NodeJS.Timeout | null>(null);

  /**
   * 获取容器元素
   */
  const getContainer = useCallback((): HTMLElement | null => {
    if (!enabled) return null;

    // 如果已经缓存了容器，直接返回（前提是容器仍在 DOM 中）
    if (containerRef.current && document.contains(containerRef.current)) {
      return containerRef.current;
    }

    // 根据不同的 container 类型获取元素
    let element: HTMLElement | null = null;

    if (typeof container === 'string') {
      element = document.querySelector<HTMLElement>(container);
    } else if (typeof container === 'function') {
      element = container();
    } else if (container instanceof HTMLElement) {
      element = container;
    }

    // 缓存容器引用
    if (element) {
      containerRef.current = element;
    }

    return element;
  }, [container, enabled]);

  /**
   * 检查用户是否在底部（只读检查，不修改状态）
   */
  const checkIfAtBottom = useCallback((): boolean => {
    const element = getContainer();
    if (!element) return false;

    const isAtBottom =
      element.scrollHeight - element.scrollTop - element.clientHeight <=
      threshold;

    return isAtBottom;
  }, [getContainer, threshold]);

  /**
   * 处理滚轮事件 - 判断滚动方向
   * 只有向上滚动且不在底部时才禁用自动滚动
   */
  const handleWheel = useCallback(
    (event: WheelEvent) => {
      if (!enabled) return;

      const element = getContainer();
      if (!element) return;

      // deltaY > 0 表示向下滚动，< 0 表示向上滚动
      const isScrollingUp = event.deltaY < 0;

      // 只有向上滚动且不在底部时才禁用自动滚动
      if (isScrollingUp) {
        userInteractingRef.current = true;
        shouldAutoScrollRef.current = false;
        setShouldAutoScroll(false);

        // 清除之前的恢复计时器
        if (resumeTimerRef.current) {
          clearTimeout(resumeTimerRef.current);
          resumeTimerRef.current = null;
        }
      }
    },
    [enabled, getContainer],
  );

  /**
   * 处理触摸/点击事件 - 任何触摸或点击滚动条都视为用户主动操作
   */
  const handleUserInteraction = useCallback(() => {
    if (!enabled) return;

    const element = getContainer();
    if (!element) return;

    // 检查是否在底部阈值内
    const distanceFromBottom =
      element.scrollHeight - element.scrollTop - element.clientHeight;
    const isAtBottom = distanceFromBottom <= threshold;

    // 如果不在底部，才禁用自动滚动
    if (!isAtBottom) {
      userInteractingRef.current = true;
      shouldAutoScrollRef.current = false;
      setShouldAutoScroll(false);

      // 清除之前的恢复计时器
      if (resumeTimerRef.current) {
        clearTimeout(resumeTimerRef.current);
        resumeTimerRef.current = null;
      }
    }
  }, [enabled, threshold, getContainer]);

  /**
   * 处理滚动事件
   * 仅用于检测用户是否滚动到底部，以便恢复自动滚动
   */
  const handleScrollEvent = useCallback(() => {
    if (!enabled) return;

    // 清除之前的恢复计时器
    if (resumeTimerRef.current) {
      clearTimeout(resumeTimerRef.current);
      resumeTimerRef.current = null;
    }

    // 使用防抖检查是否在底部
    resumeTimerRef.current = setTimeout(() => {
      const element = getContainer();
      if (!element) return;

      const scrollTop = element.scrollTop;
      const scrollHeight = element.scrollHeight;
      const clientHeight = element.clientHeight;
      const distanceFromBottom = scrollHeight - scrollTop - clientHeight;
      const isAtBottom = distanceFromBottom <= threshold;

      // 如果用户滚动到底部，恢复自动滚动
      if (isAtBottom && !shouldAutoScrollRef.current) {
        userInteractingRef.current = false;
        shouldAutoScrollRef.current = true;
        setShouldAutoScroll(true);
      }
    }, resumeDebounceMs);
  }, [enabled, threshold, resumeDebounceMs, getContainer]);

  /**
   * 强制滚动到底部（忽略 shouldAutoScroll 状态，并重置为允许自动滚动）
   */
  const forceScrollToBottom = useCallback(() => {
    if (!enabled) return;

    const element = getContainer();
    if (element) {
      // 强制滚动时，重置为允许自动滚动状态
      userInteractingRef.current = false;
      shouldAutoScrollRef.current = true;
      setShouldAutoScroll(true);

      // 清除恢复计时器
      if (resumeTimerRef.current) {
        clearTimeout(resumeTimerRef.current);
        resumeTimerRef.current = null;
      }

      element.scrollTo({
        top: element.scrollHeight,
        behavior,
      });
    }
  }, [getContainer, behavior, enabled]);

  /**
   * 滚动到底部（会根据 shouldAutoScroll 判断）
   * 注意：这里使用 ref 而不是 state，确保检查的是最新的同步值
   */
  const scrollToBottom = useCallback(() => {
    if (!shouldAutoScrollRef.current || !enabled) return;
    forceScrollToBottom();
  }, [forceScrollToBottom, enabled]);

  /**
   * 监听用户交互事件和滚动事件
   */
  useEffect(() => {
    if (!enabled) return;

    const element = getContainer();
    if (!element) return;

    // 监听用户交互事件（表明用户主动操作）
    element.addEventListener('wheel', handleWheel as EventListener, {
      passive: true,
    });
    element.addEventListener('touchstart', handleUserInteraction, {
      passive: true,
    });
    // 监听滚动事件（用于检测是否回到底部）
    element.addEventListener('scroll', handleScrollEvent, { passive: true });

    return () => {
      element.removeEventListener('wheel', handleWheel as EventListener);
      element.removeEventListener('touchstart', handleUserInteraction);
      element.removeEventListener('scroll', handleScrollEvent);

      // 清理恢复计时器
      if (resumeTimerRef.current) {
        clearTimeout(resumeTimerRef.current);
        resumeTimerRef.current = null;
      }
    };
  }, [
    getContainer,
    handleScrollEvent,
    handleWheel,
    handleUserInteraction,
    enabled,
  ]);

  /**
   * 监听容器内容高度变化（使用 ResizeObserver）
   * 当内容高度增加且允许自动滚动时，自动滚动到底部
   */
  useEffect(() => {
    if (!enabled) return;

    const element = getContainer();
    if (!element) return;

    // 获取滚动容器的第一个子元素（实际包含内容的元素）
    const contentElement = element.firstElementChild as HTMLElement;
    if (!contentElement) return;

    // 使用 ResizeObserver 监听内容元素的尺寸变化
    const resizeObserver = new ResizeObserver(() => {
      // 只有在允许自动滚动时才触发
      if (shouldAutoScrollRef.current) {
        // 使用 requestAnimationFrame 确保在 DOM 更新后滚动
        requestAnimationFrame(() => {
          scrollToBottom();
        });
      }
    });

    resizeObserver.observe(contentElement);

    return () => {
      resizeObserver.disconnect();
    };
  }, [enabled, getContainer, scrollToBottom]);

  /**
   * 手动设置是否应该自动滚动（包装函数，同时更新 state 和 ref）
   */
  const setShouldAutoScrollWrapper = useCallback((value: boolean) => {
    shouldAutoScrollRef.current = value;
    setShouldAutoScroll(value);

    // 如果设置为 true，重置用户交互状态
    if (value) {
      userInteractingRef.current = false;

      // 清除恢复计时器
      if (resumeTimerRef.current) {
        clearTimeout(resumeTimerRef.current);
        resumeTimerRef.current = null;
      }
    }
  }, []);

  return {
    shouldAutoScroll,
    scrollToBottom,
    forceScrollToBottom,
    setShouldAutoScroll: setShouldAutoScrollWrapper,
    checkIfAtBottom,
    getContainer,
  };
}
