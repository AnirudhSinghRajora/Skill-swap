'use client';

import { useEffect, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { ArrowUpRight } from 'lucide-react';

/* Two columns of skills that rotate in tandem.
   Index N on the left "trades" with index N on the right.
   The middle SVG arc connects them visually. */
const SWAPS: Array<[string, string]> = [
	['Photography', 'Spanish'],
	['React', 'Portrait lighting'],
	['Watercolor', 'Public speaking'],
	['Mandarin', 'Guitar'],
	['Pottery', 'Calculus'],
	['Bread baking', 'UX research']
];

const HeroSection = () => {
	const [i, setI] = useState(0);

	useEffect(() => {
		const t = setInterval(() => setI((n) => (n + 1) % SWAPS.length), 2600);
		return () => clearInterval(t);
	}, []);

	const [left, right] = SWAPS[i];

	return (
		<section className="relative overflow-hidden">
			{/* Top hairline rule — editorial framing */}
			<div className="absolute inset-x-0 top-0 h-px bg-foreground/10" />

			<div className="mx-auto max-w-[1400px] px-6 pt-24 pb-28 sm:px-10 md:pt-32 md:pb-36 lg:pt-40 lg:pb-44">
				{/* Folio header — magazine masthead */}
				<motion.div
					initial={{ opacity: 0, y: 8 }}
					animate={{ opacity: 1, y: 0 }}
					transition={{ duration: 0.5, ease: [0.22, 1, 0.36, 1] }}
					className="mb-16 flex items-end justify-between gap-6 border-b border-foreground/10 pb-6 md:mb-24"
				>
					<div className="flex items-center gap-4">
						<span className="folio">№ 01</span>
						<div className="hidden flex-col text-[0.7rem] leading-tight tracking-wider text-muted-foreground sm:flex">
							<span className="uppercase">Volume One</span>
							<span className="italic">A field guide to swapping skills</span>
						</div>
					</div>
					<div className="hidden text-right text-[0.7rem] uppercase leading-tight tracking-[0.2em] text-muted-foreground md:block">
						Est. 2024 &middot; For curious people
					</div>
				</motion.div>

				{/* Headline — left-weighted, asymmetric */}
				<div className="grid gap-12 lg:grid-cols-12 lg:gap-16">
					<div className="lg:col-span-9">
						<motion.h1
							initial={{ opacity: 0, y: 14 }}
							animate={{ opacity: 1, y: 0 }}
							transition={{ duration: 0.7, ease: [0.22, 1, 0.36, 1], delay: 0.1 }}
							className="text-display-xl text-foreground"
						>
							Trade your{' '}
							<SwappingWord word={left} side="left" />
							<br />
							for someone&rsquo;s{' '}
							<SwappingWord word={right} side="right" />.
						</motion.h1>
					</div>

					{/* Right column — sparse editorial intro */}
					<motion.div
						initial={{ opacity: 0, y: 14 }}
						animate={{ opacity: 1, y: 0 }}
						transition={{ duration: 0.7, ease: [0.22, 1, 0.36, 1], delay: 0.25 }}
						className="lg:col-span-3 lg:pt-8"
					>
						<p className="eyebrow mb-4">The pitch</p>
						<p className="font-serif text-[1.15rem] leading-snug text-foreground/85 italic">
							No subscriptions. No middlemen. Just two people, one teaching what they
							love and learning what they need.
						</p>
					</motion.div>
				</div>

				{/* The connecting arc — hero moment.
				    Hand-drawn SVG path that runs from the left swap word
				    down and across to the right swap word, drawing on mount. */}
				<motion.div
					initial={{ opacity: 0 }}
					animate={{ opacity: 1 }}
					transition={{ duration: 0.6, delay: 0.4 }}
					className="pointer-events-none mt-6 hidden h-24 lg:block"
					aria-hidden
				>
					<svg
						viewBox="0 0 1200 120"
						className="h-full w-full"
						fill="none"
						stroke="currentColor"
						strokeWidth="1.5"
						strokeLinecap="round"
						strokeLinejoin="round"
					>
						<path
							className="animate-draw text-primary"
							style={{ ['--draw-length' as string]: '1400' }}
							d="M 120 12 C 220 60, 360 95, 540 80 S 880 30, 1080 70"
						/>
						{/* arrow head */}
						<path
							className="animate-draw text-primary"
							style={{
								['--draw-length' as string]: '40',
								animationDelay: '1.6s'
							}}
							d="M 1058 60 L 1080 70 L 1066 88"
						/>
					</svg>
				</motion.div>

				{/* Action row — single primary, single quiet secondary, three editorial stats */}
				<motion.div
					initial={{ opacity: 0, y: 12 }}
					animate={{ opacity: 1, y: 0 }}
					transition={{ duration: 0.6, ease: [0.22, 1, 0.36, 1], delay: 0.5 }}
					className="mt-12 flex flex-col items-start justify-between gap-10 border-t border-foreground/10 pt-10 lg:flex-row lg:items-end lg:gap-16"
				>
					<div className="flex flex-wrap items-center gap-3">
						<Button asChild variant="brand" size="lg">
							<Link href="/auth/signup">
								Start a swap
								<ArrowUpRight className="ml-1 h-4 w-4" />
							</Link>
						</Button>
						<Button asChild variant="link" size="lg">
							<Link href="/browse" className="link-sweep">
								Or browse the directory
							</Link>
						</Button>
					</div>

					<dl className="grid grid-cols-3 gap-x-10 gap-y-1 text-sm sm:gap-x-16">
						<HeroStat value="12k" label="Swappers" />
						<HeroStat value="64" label="Countries" />
						<HeroStat value="4.9" label="Avg. session" />
					</dl>
				</motion.div>
			</div>
		</section>
	);
};

function SwappingWord({ word, side }: { word: string; side: 'left' | 'right' }) {
	return (
		<span className="relative inline-block align-baseline">
			<AnimatePresence mode="popLayout" initial={false}>
				<motion.span
					key={word}
					initial={{ y: '60%', opacity: 0, rotate: side === 'left' ? -2 : 2 }}
					animate={{ y: 0, opacity: 1, rotate: 0 }}
					exit={{ y: '-60%', opacity: 0, rotate: side === 'left' ? 2 : -2 }}
					transition={{ duration: 0.55, ease: [0.22, 1, 0.36, 1] }}
					className="inline-block italic text-primary"
				>
					{word}
				</motion.span>
			</AnimatePresence>
		</span>
	);
}

function HeroStat({ value, label }: { value: string; label: string }) {
	return (
		<div className="flex flex-col">
			<dt className="font-serif text-3xl leading-none text-foreground sm:text-4xl">
				{value}
			</dt>
			<dd className="mt-1 text-xs uppercase tracking-[0.18em] text-muted-foreground">
				{label}
			</dd>
		</div>
	);
}

export default HeroSection;
