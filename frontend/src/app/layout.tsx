import type { Metadata } from 'next';
import './globals.css';
import { ErrorProvider } from '@/components/ErrorContext';
import ToastContainer from '@/components/ToastContainer';

export const metadata: Metadata = {
  title: 'Yt2bulk - YouTube Export & Intelligence',
  description: 'High-performance YouTube metadata, transcript, and signal extraction toolkit',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className="bg-[#050505] text-[#f0f0f0] min-h-screen antialiased">
        <ErrorProvider>
          {children}
          <ToastContainer />
        </ErrorProvider>
      </body>
    </html>
  );
}
