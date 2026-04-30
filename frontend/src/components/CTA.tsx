'use client';

import { useRef } from 'react';
import { motion, useInView } from 'framer-motion';
import { Button } from '@/components/ui/button';
import { ArrowUpRight } from 'lucide-react';
import Link from 'next/link';

const CallToActionSection = () => {
	const ref = useRef(null);
	const inView = useInView(ref, { once: true, amount: 0.2 });

	return (
		<section
			ref={ref}
			className="relative overflow-hidden border-t border-foreground/10 bg-foreground py-32 text-background md:py-48"
		>
			{/* A single ruled hairline above echoes the magazine framing */}
			<div className="absolute inset-x-0 top-0 h-px bg-background/15" />

			<div className="mx-auto max-w-[1400px] px-6 sm:px-10">
				{/* Footer-folio */}
				<motion.div
					initial={{ opacity: 0, y: 8 }}
					animate={inView ? { opacity: 1, y: 0 } : {}}
					transition={{ duration: 0.6, ease: [0.22, 1, 0.36, 1] }}
					className="mb-16 flex items-baseline justify-between border-b border-background/15 pb-4 md:mb-24"
				>
					<span className="folio !text-background/60">№ 05</span>
					<span className="text-[0.7rem] uppercase tracking-[0.2em] text-background/55">
						The closing argument
					</span>
				</motion.div>

				{/* Closing headline — typography as the entire layout */}
				<motion.h2
					initial={{ opacity: 0, y: 16 }}
					animate={inView ? { opacity: 1, y: 0 } : {}}
					transition={{ duration: 0.8, ease: [0.22, 1, 0.36, 1] }}
					className="text-display-xl text-background"
				>
					Somewhere<br />
					a stranger knows<br />
					exactly what you want{' '}
					<em className="text-primary">to learn.</em>
				</motion.h2>

				<motion.p
					initial={{ opacity: 0, y: 12 }}
					animate={inView ? { opacity: 1, y: 0 } : {}}
					transition={{ duration: 0.6, delay: 0.2, ease: [0.22, 1, 0.36, 1] }}
					className="mt-12 max-w-[44ch] font-serif text-xl italic leading-snug text-background/75 md:text-2xl"
				>
					And, surprisingly often, they happen to want
					to learn what only you can teach.
				</motion.p>

				{/* CTA — single brand button, asymmetric, left-aligned */}
				<motion.div
					initial={{ opacity: 0, y: 12 }}
					animate={inView ? { opacity: 1, y: 0 } : {}}
					transition={{ duration: 0.6, delay: 0.35, ease: [0.22, 1, 0.36, 1] }}
					className="mt-16 flex flex-col items-start gap-8 border-t border-background/15 pt-10 md:flex-row md:items-end md:justify-between"
				>
					<div className="flex flex-wrap items-center gap-4">
						<Button asChild variant="brand" size="lg">
							<Link href="/auth/signup">
								Find them
								<ArrowUpRight className="ml-1 h-4 w-4" />
							</Link>
						</Button>
						<Link
							href="/auth/signin"
							className="link-sweep text-sm text-background/70 hover:text-background"
						>
							Already a member? Sign in.
						</Link>
					</div>

					<p className="font-serif text-sm italic text-background/55">
						SkillSwap, Volume One &middot; Made for the curious.
					</p>
				</motion.div>
			</div>
		</section>
	);
};

export default CallToActionSection;
