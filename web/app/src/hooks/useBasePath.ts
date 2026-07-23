import { useStore } from '@/provider';
import { getBasePath } from '@/utils';
export const useBasePath = () => {
  const { kbDetail } = useStore();
  const url = kbDetail?.base_url;
  if (!url) return '';
  return getBasePath(url);
};
