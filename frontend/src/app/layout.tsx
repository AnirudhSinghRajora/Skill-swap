import type { Metadata } from 'next';
import { Geist, Geist_Mono, Instrument_Serif } from 'next/font/google';
import { QueryClientProvider } from '@/providers/QueryClientProvider';
import { ThemeProvider } from '@/providers/ThemeProvider';
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

const instrumentSerif = Instrument_Serif({
	variable: '--font-instrument-serif',
	subsets: ['latin'],
	weight: '400',
	style: ['normal', 'italic'],
	display: 'swap'
});

export const metadata: Metadata = {
	title: 'SkillSwap — Trade what you know for what you want to learn',
	description:
		'A peer-to-peer skill exchange for curious people. Teach what you love, learn what you need. No fees, no middlemen — just trade.'
};

export default function RootLayout({
	children
}: Readonly<{
	children: React.ReactNode;
}>) {
	return (
		<html lang="en" suppressHydrationWarning>
			<body
				className={`${geistSans.variable} ${geistMono.variable} ${instrumentSerif.variable} antialiased bg-background font-sans`}
			>
				<ThemeProvider>
					<SmoothScroll />
					<QueryClientProvider>
						<Navigation />
						<main className="min-h-screen">{children}</main>
						<Toaster richColors position="top-right" />
					</QueryClientProvider>
				</ThemeProvider>
			</body>
		</html>
	);
}
