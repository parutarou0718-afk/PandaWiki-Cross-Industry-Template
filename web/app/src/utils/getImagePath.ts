export const getImagePath = (path: string, basePath?: string) => {
  if (!path) return path;
  if (path.startsWith('http') || path.startsWith('blob')) {
    return path;
  }
  const basePathValue =
    basePath || (typeof window !== 'undefined' ? window._BASE_PATH_ || '' : '');
  if (path.startsWith(basePathValue as string)) {
    return path;
  }
  return `${basePathValue}${path}`;
};
