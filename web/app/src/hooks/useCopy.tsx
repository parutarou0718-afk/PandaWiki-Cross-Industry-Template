'use client';

import { useCallback, useEffect } from 'react';

export type CopyControlMode = 'disable' | 'allow';

export interface UseCopyOptions {
  mode?: CopyControlMode;
  suffix?: string;
  /**
   * 绑定事件的目标容器。
   * - 不传：默认绑定到 document
   * - 传入元素：仅在该元素范围内拦截事件
   */
  target?: Document | HTMLElement | null;
  /**
   * 当禁用复制时，是否同时禁用右键菜单（防止通过右键菜单复制）。默认 true
   */
  blockContextMenuWhenDisabled?: boolean;
}

export interface UseCopyReturn {
  /** 程序化复制文本（自动追加 suffix）。禁用模式下返回 false */
  copy: (text: string) => Promise<boolean>;
}

/**
 * 控制复制行为的 Hook：
 * - mode: "disable" 全局（或目标内）禁用复制（含快捷键与右键菜单）
 * - mode: "allow" 允许复制；如提供 suffix，则在复制内容末尾追加
 */
export function useCopy(options: UseCopyOptions = {}): UseCopyReturn {
  const {
    mode = 'allow',
    suffix,
    target = typeof document !== 'undefined' ? document : null,
    blockContextMenuWhenDisabled = true,
  } = options;

  useEffect(() => {
    if (!target) return;

    const onBeforeCopy = (e: Event) => {
      if (mode === 'disable') {
        e.preventDefault();
      }
    };

    const onBeforeCut = (e: Event) => {
      if (mode === 'disable') {
        e.preventDefault();
      }
    };

    const onCopy = (e: ClipboardEvent) => {
      if (mode === 'disable') {
        e.preventDefault();
        return;
      }
      if (!suffix) return;

      try {
        // 读取当前选择文本
        const selection =
          typeof window !== 'undefined' && window.getSelection
            ? (window.getSelection()?.toString() ?? '')
            : '';
        const originalText =
          selection || e.clipboardData?.getData('text/plain') || '';
        const appended = originalText + suffix;
        if (e.clipboardData) {
          e.clipboardData.setData('text/plain', appended);
          // 尝试同时更新 HTML（简单处理，尾缀作为纯文本追加）
          const originalHtml = e.clipboardData.getData('text/html');
          if (originalHtml) {
            const appendedHtml = originalHtml + suffix;
            e.clipboardData.setData('text/html', appendedHtml);
          }
          // 阻止默认，让我们设置的内容生效
          e.preventDefault();
        }
      } catch {}
    };

    const onCut = (e: ClipboardEvent) => {
      if (mode === 'disable') {
        e.preventDefault();
        return;
      }
      if (!suffix) return;
      try {
        const selection =
          typeof window !== 'undefined' && window.getSelection
            ? (window.getSelection()?.toString() ?? '')
            : '';
        const originalText =
          selection || e.clipboardData?.getData('text/plain') || '';
        const appended = originalText + suffix;
        if (e.clipboardData) {
          e.clipboardData.setData('text/plain', appended);
          const originalHtml = e.clipboardData.getData('text/html');
          if (originalHtml) {
            const appendedHtml = originalHtml + suffix;
            e.clipboardData.setData('text/html', appendedHtml);
          }
          e.preventDefault();
        }
      } catch {}
    };

    const onKeyDown = (e: KeyboardEvent) => {
      if (mode !== 'disable') return;
      const isCopyShortcut =
        (e.metaKey || e.ctrlKey) && (e.key === 'c' || e.key === 'C');
      const isCutShortcut =
        (e.metaKey || e.ctrlKey) && (e.key === 'x' || e.key === 'X');
      const isCopyInsert = e.ctrlKey && e.key === 'Insert';
      if (isCopyShortcut || isCutShortcut || isCopyInsert) {
        e.preventDefault();
      }
    };

    const onContextMenu = (e: MouseEvent) => {
      if (mode === 'disable' && blockContextMenuWhenDisabled) {
        e.preventDefault();
      }
    };

    // 事件绑定（元素或 document）
    const add = (
      name: string,
      handler: EventListenerOrEventListenerObject,
      opts?: boolean | AddEventListenerOptions,
    ) => {
      (target as any).addEventListener(name, handler, opts);
    };
    const remove = (
      name: string,
      handler: EventListenerOrEventListenerObject,
      opts?: boolean | EventListenerOptions,
    ) => {
      (target as any).removeEventListener(name, handler, opts);
    };

    add('beforecopy', onBeforeCopy as EventListener, true);
    add('beforecut', onBeforeCut as EventListener, true);
    add('copy', onCopy as EventListener, true);
    add('cut', onCut as EventListener, true);
    add('keydown', onKeyDown as EventListener);
    add('contextmenu', onContextMenu as EventListener, true);

    return () => {
      remove('beforecopy', onBeforeCopy as EventListener, true);
      remove('beforecut', onBeforeCut as EventListener, true);
      remove('copy', onCopy as EventListener, true);
      remove('cut', onCut as EventListener, true);
      remove('keydown', onKeyDown as EventListener);
      remove('contextmenu', onContextMenu as EventListener, true);
    };
  }, [mode, suffix, target, blockContextMenuWhenDisabled]);

  const copy = useCallback(
    async (text: string) => {
      if (mode === 'disable') return false;
      const payload = suffix ? text + suffix : text;
      try {
        if (navigator.clipboard && navigator.clipboard.writeText) {
          await navigator.clipboard.writeText(payload);
          return true;
        }
        // 旧浏览器回退
        const textarea = document.createElement('textarea');
        textarea.value = payload;
        document.body.appendChild(textarea);
        textarea.select();
        document.execCommand('copy');
        document.body.removeChild(textarea);
        return true;
      } catch {
        return false;
      }
    },
    [mode, suffix],
  );

  return { copy };
}

export default useCopy;
