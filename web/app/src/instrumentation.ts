import * as Sentry from '@sentry/nextjs';

export async function register() {
  // 只在生产环境下启用 Sentry
  if (process.env.NODE_ENV === 'production') {
    if (process.env.NEXT_RUNTIME === 'nodejs') {
      await import('../sentry.server.config');
    }

    if (process.env.NEXT_RUNTIME === 'edge') {
      await import('../sentry.edge.config');
    }
  }
}

// 只在生产环境下导出错误捕获函数
export const onRequestError =
  process.env.NODE_ENV === 'production'
    ? Sentry.captureRequestError
    : undefined;
