'use client';

import { motion, useInView } from 'framer-motion';
import { useRef } from 'react';
import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { ArrowUpRight } from 'lucide-react';

const SKILL_LANES: Array<{ teaching: string[]; learning: string[] }> = [
	{
		teaching: [
			'React performance',
			'Portrait lighting',
			'Sourdough',
			'Mandarin tones',
			'Watercolor washes',
			'TypeScript types',
			'Negotiation',
			'Cold-process soap',
			'Jazz piano',
			'Public speaking',
			'Video editing',
			'Spanish',
			'Calligraphy',
			'Rust',
			'Climbing technique',
			'Improv',
			'UX research'
		],
		learning: [
			'Calculus',
			'Graphic design',
			'Bouldering',
			'French',
			'Pottery',
			'Branding',
			'SQL',
			'Yoga',
			'Crochet',
			'Filmmaking',
			'Investing',
			'Korean',
			'Animation',
			'Bartending',
			'Drawing hands',
			'Poetry',
			'Cooking Thai'
		]
	}
];

const SkillsMarketplace = () => {
	const ref = useRef(null);
	const inView = useInView(ref, { once: true, amount: 0.15 });

	const lane = SKILL_LANES[0];

	return (
		<section ref={ref} className="relative border-t border-foreground/10 bg-background py-28 md:py-40">
			<div className="mx-auto max-w-[1400px] px-6 sm:px-10">
				{/* Asymmetric editorial header */}
				<div className="mb-20 grid gap-10 md:grid-cols-12 md:gap-16">
					<motion.div
						initial={{ opacity: 0, y: 12 }}
						animate={inView ? { opacity: 1, y: 0 } : {}}
						transition={{ duration: 0.6, ease: [0.22, 1, 0.36, 1] }}
						className="md:col-span-5"
					>
						<div className="flex items-baseline gap-4">
							<span className="folio">№ 02</span>
							<p className="eyebrow">The Wire</p>
						</div>
						<h2 className="text-display-lg mt-6 text-foreground">
							A directory<br />
							<em className="text-primary">always in motion.</em>
						</h2>
					</motion.div>

					<motion.div
						initial={{ opacity: 0, y: 12 }}
						animate={inView ? { opacity: 1, y: 0 } : {}}
						transition={{ duration: 0.6, delay: 0.1, ease: [0.22, 1, 0.36, 1] }}
						className="md:col-span-6 md:col-start-7 md:pt-12"
					>
						<p className="font-serif text-xl leading-snug text-foreground/85 md:text-2xl">
							Every minute, somebody offers to teach
							you something you couldn&rsquo;t learn from a course.
							Slow down to read it &mdash; or hover the wire to pause.
						</p>
					</motion.div>
				</div>

				{/* Two-row wire ticker — opposing directions for visual rhythm.
				    Typography IS the layout. */}
				<motion.div
					initial={{ opacity: 0 }}
					animate={inView ? { opacity: 1 } : {}}
					transition={{ duration: 0.8, delay: 0.2 }}
					className="-mx-6 space-y-2 sm:-mx-10"
				>
					<TickerRow
						label="Offering"
						items={lane.teaching}
						direction="left"
						accent
					/>
					<TickerRow
						label="Wanting"
						items={lane.learning}
						direction="right"
					/>
				</motion.div>

				{/* Stats row + entry point */}
				<motion.div
					initial={{ opacity: 0, y: 12 }}
					animate={inView ? { opacity: 1, y: 0 } : {}}
					transition={{ duration: 0.6, delay: 0.3, ease: [0.22, 1, 0.36, 1] }}
					className="mt-20 grid gap-12 border-t border-foreground/10 pt-12 md:grid-cols-12 md:items-end"
				>
					<dl className="md:col-span-7 grid grid-cols-3 gap-8">
						<Counter n="805" label="Active swappers" />
						<Counter n="42" label="New today" />
						<Counter n="129" label="Skill categories" />
					</dl>

					<div className="md:col-span-5 md:text-right">
						<Button asChild variant="outline" size="lg">
							<Link href="/browse">
								Open the directory
								<ArrowUpRight className="ml-1 h-4 w-4" />
							</Link>
						</Button>
					</div>
				</motion.div>
			</div>
		</section>
	);
};

function TickerRow({
	label,
	items,
	direction,
	accent = false
}: {
	label: string;
	items: string[];
	direction: 'left' | 'right';
	accent?: boolean;
}) {
	const doubled = [...items, ...items];

	return (
		<div className="group relative flex items-center gap-4 border-y border-foreground/10 py-3 sm:py-4">
			{/* Sticky label that anchors the lane */}
			<span className="sticky left-0 z-10 flex h-full items-center bg-background pr-4 pl-6 sm:pl-10">
				<span className="eyebrow whitespace-nowrap">{label} ↦</span>
			</span>

			<div className="marquee flex-1">
				<div
					className="marquee__track gap-10 pr-10"
					style={{
						animationDirection: direction === 'right' ? 'reverse' : 'normal',
						['--marquee-duration' as string]: items.length > 12 ? '60s' : '40s'
					}}
				>
					{doubled.map((s, i) => (
						<span
							key={`${s}-${i}`}
							className={[
								'font-serif whitespace-nowrap text-3xl leading-none transition-colors duration-200 sm:text-4xl md:text-5xl',
								accent
									? 'text-foreground hover:text-primary'
									: 'text-foreground/55 italic hover:text-foreground'
							].join(' ')}
						>
							{s}
							<span className="ml-10 select-none text-foreground/15">·</span>
						</span>
					))}
				</div>
			</div>
		</div>
	);
}

function Counter({ n, label }: { n: string; label: string }) {
	return (
		<div>
			<div className="font-serif text-4xl leading-none text-foreground sm:text-5xl">
				{n}
			</div>
			<div className="mt-2 text-xs uppercase tracking-[0.18em] text-muted-foreground">
				{label}
			</div>
		</div>
	);
}

export default SkillsMarketplace;
