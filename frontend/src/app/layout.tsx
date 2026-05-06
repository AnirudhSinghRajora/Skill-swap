import type { Metadata } from 'next';
import { Geist, Geist_Mono } from 'next/font/google';
import { QueryClientProvider } from '@/providers/QueryClientProvider';
import { Navigation } from '@/components/Navigation';
import './globals.css';
import { SmoothScroll } from '@/components/smooth-scroll';
import { Toaster } from 'sonner';

const geistSans = Geist({
	variable: '--font-geist-sans',
	subsets: ['latin']
});

const geistMono = Geist_Mono({
	variable: '--font-geist-mono',
	subsets: ['latin']
});

export const metadata: Metadata = {
	title: 'SkillSwap - Exchange Skills with Others',
	description: 'Connect with people in your area to exchange skills and knowledge'
};

export default function RootLayout({
	children
}: Readonly<{
	children: React.ReactNode;
}>) {
	return (
		<html lang="en">
			<body
				className={`${geistSans.variable} ${geistMono.variable} antialiased bg-background`}
			>
				<SmoothScroll />
				<QueryClientProvider>
					<Navigation />
					<main className="min-h-screen">{children}</main>
					<Toaster richColors position="top-right" />
				</QueryClientProvider>
			</body>
		</html>
	);
}
