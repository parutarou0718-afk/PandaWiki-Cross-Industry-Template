import { ChunkResultItem } from '@/assets/type';

export interface ConversationItem {
  q: string;
  a: string;
  score: number;
  update_time: string;
  message_id: string;
  source: 'history' | 'chat';
  chunk_result: ChunkResultItem[];
  thinking_content: string;
}

export interface UploadedImage {
  id: string;
  url: string;
  file: File;
}

export interface SSEMessageData {
  type: string;
  content: string;
  chunk_result: ChunkResultItem;
}

export interface ChatRequestData {
  message: string;
  nonce: string;
  conversation_id: string;
  app_type: number;
  captcha_token: string;
}
