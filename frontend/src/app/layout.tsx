import '@mantine/core/styles.css';
import { ColorSchemeScript, MantineProvider } from '@mantine/core';

export const metadata = {
  title: 'ペンライトクイズ v2',
  description: '坂道シリーズ 公式ペンライトカラー暗記・復習クイズ',
};

export const viewport = {
  width: 'device-width',
  initialScale: 1,
  maximumScale: 1,
  userScalable: false,
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="ja" suppressHydrationWarning>
      <head>
        <ColorSchemeScript defaultColorScheme="light" />
      </head>
      <body style={{ margin: 0, padding: 0, overflowX: 'hidden' }}>
        <MantineProvider defaultColorScheme="light">{children}</MantineProvider>
      </body>
    </html>
  );
}
