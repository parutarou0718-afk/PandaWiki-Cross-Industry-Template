import WaterMarkProvider from '@/components/watermark/WaterMarkProvider';

const Layout = ({ children }: { children: React.ReactNode }) => {
  return <WaterMarkProvider>{children}</WaterMarkProvider>;
};

export default Layout;
