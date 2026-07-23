

import {  RequestParams } from "./httpClient";
import {
  GithubComChaitinPandaWikiProApiShareV1FileUploadResp,
  PostShareProV1FileUploadPayload,
} from "./types";



/**
 * 使用 XMLHttpRequest 实现文件上传进度
 */
export const postShareProV1FileUploadWithProgress = (
  data: PostShareProV1FileUploadPayload,
  params: RequestParams & {
    onprogress?: (progress: { progress: number }) => void;
    abortSignal?: AbortSignal;
  } = {},
): Promise<GithubComChaitinPandaWikiProApiShareV1FileUploadResp> => {
  return new Promise((resolve, reject) => {
    const { onprogress, abortSignal, ...requestParams } = params;
    
    // 创建 FormData
    const formData = new FormData();
    Object.keys(data).forEach(key => {
      const value = data[key as keyof PostShareProV1FileUploadPayload];
      if (value instanceof File) {
        formData.append(key, value);
      } else if (value !== null && value !== undefined) {
        formData.append(key, String(value));
      }
    });

    const xhr = new XMLHttpRequest();
    
    // 设置上传进度监听
    xhr.upload.addEventListener('progress', (event) => {
      if (event.lengthComputable && onprogress) {
        const progress = (event.loaded / event.total) * 100;
        onprogress({ progress });
      }
    });

    // 设置响应处理
    xhr.addEventListener('load', () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        try {
          const response = JSON.parse(xhr.responseText);
          if (response.code === 0 || response.code === undefined) {
            resolve(response.data);
          } else {
            reject(new Error(response.message || '上传失败'));
          }
        } catch (error) {
          reject(new Error('响应解析失败'));
        }
      } else {
        reject(new Error(`HTTP ${xhr.status}: ${xhr.statusText}`));
      }
    });

    // 设置错误处理
    xhr.addEventListener('error', () => {
      reject(new Error('网络错误'));
    });

    // 设置中止处理
    xhr.addEventListener('abort', () => {
      reject(new Error('请求被中止'));
    });

    // 监听中止信号
    if (abortSignal) {
      abortSignal.addEventListener('abort', () => {
        xhr.abort();
      });
    }

    // 构建请求 URL
    const baseUrl = process.env.TARGET || (typeof window !== 'undefined' ? window._BASE_PATH_ : '');
    const url = `${baseUrl}/share/pro/v1/file/upload`;
    
    // 发送请求
    xhr.open('POST', url);
    
    // 设置请求头
    if (requestParams.headers) {
      Object.entries(requestParams.headers).forEach(([key, value]) => {
        if (typeof value === 'string') {
          xhr.setRequestHeader(key, value);
        }
      });
    }
    
    // 设置凭据
    if (requestParams.credentials) {
      xhr.withCredentials = requestParams.credentials === 'include';
    }
    
    xhr.send(formData);
  });
};
