export const getBasePath = (path: string) => {
  if (!path) return '';
  let basePath = '';
  try {
    const u = new URL(path);
    basePath = u.pathname.replace(/\/$/, '');
  } catch {
    basePath = path.startsWith('/') ? path : `/${path}`;
  }
  return basePath;
};
