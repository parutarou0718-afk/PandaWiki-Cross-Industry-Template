// 常量定义
export const MAX_IMAGES = 9;
export const MAX_IMAGE_SIZE = 10 * 1024 * 1024; // 10MB
export const CONVERSATION_MAX_HEIGHT = 'calc(100vh - 334px)';
export const FUZZY_SUGGESTIONS_LIMIT = 5;

// 回答状态
export const AnswerStatus = {
  1: '正在搜索结果...',
  2: '思考中...',
  3: '正在回答',
  4: '',
} as const;

export type AnswerStatusType = keyof typeof AnswerStatus;
