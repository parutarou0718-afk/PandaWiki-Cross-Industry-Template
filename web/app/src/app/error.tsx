'use client';
import { FooterProvider } from '@/components/footer';
import { Stack } from '@mui/material';
import ErrorComponent from '@/components/error';

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <Stack
      justifyContent='space-between'
      alignItems='center'
      sx={{
        height: '100vh',
      }}
    >
      <Stack flex={1} justifyContent='center' alignItems='center'>
        <ErrorComponent error={error} reset={reset} />
      </Stack>
      <FooterProvider showBrand={false} />
    </Stack>
  );
}
