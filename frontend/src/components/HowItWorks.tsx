'use client';

import { useRef } from 'react';
import { motion, useInView } from 'framer-motion';
import { Button } from '@/components/ui/button';
import Link from 'next/link';
import { ArrowUpRight } from 'lucide-react';

const STEPS = [
	{
		n: 'i',
		title: 'Write what you know.',
		body: 'A profile, in your own voice — what you teach, what you wish you understood, when you have an hour. No résumé fields. No skill tags pretending to be a personality test.'
	},
	{
		n: 'ii',
		title: 'Find your mirror.',
		body: 'Browse the directory or let us suggest. The good matches feel obvious — they teach what you want and want what you teach.'
	},
	{
		n: 'iii',
		title: 'Trade in private.',
		body: 'Open a chat, swap voice notes, send a draft, hop on a call. Your messages are end-to-end encrypted; no one but the two of you reads them.'
	},
	{
		n: 'iv',
		title: 'Close the loop.',
		body: 'When the swap feels done, both of you mark it complete. A short reflection, a public thank-you. Then you start the next one.'
	}
];

const HowItWorksSection = () => {
	const ref = useRef(null);
	const inView = useInView(ref, { once: true, amount: 0.1 });

	return (
		<section ref={ref} className="border-t border-foreground/10 bg-background py-28 md:py-40">
			<div className="mx-auto max-w-[1400px] px-6 sm:px-10">
				{/* Masthead */}
				<motion.div
					initial={{ opacity: 0, y: 12 }}
					animate={inView ? { opacity: 1, y: 0 } : {}}
					transition={{ duration: 0.6, ease: [0.22, 1, 0.36, 1] }}
					className="mb-24 grid gap-10 md:grid-cols-12"
				>
					<div className="md:col-span-6">
						<div className="flex items-baseline gap-4">
							<span className="folio">№ 04</span>
							<p className="eyebrow">The method</p>
						</div>
						<h2 className="text-display-lg mt-6 text-foreground">
							Four moves, then <em className="text-primary">repeat.</em>
						</h2>
					</div>

					<p className="font-serif text-xl leading-snug text-foreground/80 md:col-span-5 md:col-start-8 md:pt-12 md:text-2xl">
						The whole cycle, end to end, takes most people less than a week
						the first time and an afternoon every time after.
					</p>
				</motion.div>

				{/* Steps — alternating indent, breaking the grid */}
				<ol className="relative">
					{/* Vertical rule */}
					<div
						aria-hidden
						className="absolute left-[7%] top-0 hidden h-full w-px bg-foreground/10 md:block"
					/>

					{STEPS.map((step, i) => (
						<motion.li
							key={step.n}
							initial={{ opacity: 0, y: 16 }}
							animate={inView ? { opacity: 1, y: 0 } : {}}
							transition={{ duration: 0.65, delay: 0.1 * i, ease: [0.22, 1, 0.36, 1] }}
							className="grid grid-cols-1 gap-6 border-b border-foreground/10 py-12 md:grid-cols-12 md:gap-10 md:py-16 last:border-b-0"
						>
							<div className="md:col-span-2">
								<span className="folio italic">{step.n}.</span>
							</div>

							{/* Indent cycles 0/1/2/1 to break uniform alignment */}
							<div
								className={[
									'md:col-span-9',
									i === 1 ? 'md:col-start-4' : i === 2 ? 'md:col-start-5' : i === 3 ? 'md:col-start-4' : 'md:col-start-3'
								].join(' ')}
							>
								<h3 className="text-display-md text-foreground">{step.title}</h3>
								<p className="mt-5 max-w-[60ch] text-base leading-relaxed text-foreground/75 md:text-lg">
									{step.body}
								</p>
							</div>
						</motion.li>
					))}
				</ol>

				{/* Soft outro CTA — left-aligned, asymmetric */}
				<motion.div
					initial={{ opacity: 0, y: 12 }}
					animate={inView ? { opacity: 1, y: 0 } : {}}
					transition={{ duration: 0.6, delay: 0.4, ease: [0.22, 1, 0.36, 1] }}
					className="mt-20 flex flex-col items-start gap-4 sm:flex-row sm:items-center sm:gap-6"
				>
					<Button asChild variant="brand" size="lg">
						<Link href="/auth/signup">
							Begin your first swap
							<ArrowUpRight className="ml-1 h-4 w-4" />
						</Link>
					</Button>
					<p className="text-sm text-muted-foreground">
						Two minutes to set up. No card required.
					</p>
				</motion.div>
			</div>
		</section>
	);
};

export default HowItWorksSection;
