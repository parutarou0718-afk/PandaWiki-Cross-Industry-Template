import { ReportPage } from '@/views/reports';
export default async function ReportDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  return <ReportPage id={id} />;
}
