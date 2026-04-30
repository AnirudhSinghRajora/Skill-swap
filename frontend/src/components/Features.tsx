'use client';

import { motion } from 'framer-motion';
import { useInView } from 'react-intersection-observer';

const FEATURES = [
	{
		size: 'lead',
		eyebrow: 'The match',
		title: 'Pairing that feels almost suspiciously good.',
		body: 'You list what you teach and what you wish you knew. We surface the people whose answer is the exact mirror of yours. No algorithm theatre — just two well-formed questions meeting in the middle.'
	},
	{
		size: 'medium',
		eyebrow: 'Trust',
		title: 'Verified profiles. Reviewed sessions.',
		body: 'Every swap closes with a two-way reflection. Reputation grows the way it does in real life — slowly, in public, on the strength of work you actually did.'
	},
	{
		size: 'medium',
		eyebrow: 'Privacy',
		title: 'End-to-end encrypted chat.',
		body: 'Your messages are encrypted in your browser, before they leave your device. We can&rsquo;t read them. Neither can anyone we work with. That&rsquo;s by design, not promise.'
	},
	{
		size: 'medium',
		eyebrow: 'Live',
		title: 'Video built into the swap.',
		body: 'When you&rsquo;re ready, click into a session room. No download, no separate app. Camera, screen, and shared focus — that&rsquo;s the whole feature set.'
	},
	{
		size: 'small',
		eyebrow: 'Free',
		title: 'No fees. Ever.',
		body: 'Time is the currency. We don&rsquo;t take a cut of either side.'
	},
	{
		size: 'small',
		eyebrow: 'Pace',
		title: 'Async or live.',
		body: 'Trade voice notes at midnight or meet on Sunday morning. Both count.'
	}
] as const;

const FeaturesSection = () => {
	const [ref, inView] = useInView({ threshold: 0.08, triggerOnce: true });

	return (
		<section ref={ref} className="border-t border-foreground/10 bg-secondary/40 py-28 md:py-40">
			<div className="mx-auto max-w-[1400px] px-6 sm:px-10">
				{/* Section masthead */}
				<motion.div
					initial={{ opacity: 0, y: 12 }}
					animate={inView ? { opacity: 1, y: 0 } : {}}
					transition={{ duration: 0.6, ease: [0.22, 1, 0.36, 1] }}
					className="mb-20 grid gap-10 md:grid-cols-12 md:gap-16"
				>
					<div className="md:col-span-5">
						<div className="flex items-baseline gap-4">
							<span className="folio">№ 03</span>
							<p className="eyebrow">What&rsquo;s inside</p>
						</div>
						<h2 className="text-display-lg mt-6 text-foreground">
							Built for the<br />
							<em className="text-primary">long swap.</em>
						</h2>
					</div>

					<p className="font-serif text-xl leading-snug text-foreground/80 md:col-span-6 md:col-start-7 md:pt-12 md:text-2xl">
						Six things we got obsessive about, so you don&rsquo;t have to think about them.
						The rest of the surface area, we kept deliberately small.
					</p>
				</motion.div>

				{/* Editorial bento — 12 col, varied weights */}
				<div className="grid grid-cols-1 gap-x-12 gap-y-16 md:grid-cols-12 md:gap-y-20">
					{FEATURES.map((f, i) => (
						<FeatureBlock key={f.title} feature={f} index={i} inView={inView} />
					))}
				</div>
			</div>
		</section>
	);
};

function FeatureBlock({
	feature,
	index,
	inView
}: {
	feature: (typeof FEATURES)[number];
	index: number;
	inView: boolean;
}) {
	const span = {
		lead: 'md:col-span-12',
		medium: 'md:col-span-4',
		small: 'md:col-span-3'
	}[feature.size];

	const offset = feature.size === 'small' && index === FEATURES.length - 1
		? 'md:col-start-7'
		: feature.size === 'small'
			? 'md:col-start-4'
			: '';

	return (
		<motion.article
			initial={{ opacity: 0, y: 16 }}
			animate={inView ? { opacity: 1, y: 0 } : {}}
			transition={{ duration: 0.6, delay: 0.06 * index, ease: [0.22, 1, 0.36, 1] }}
			className={`${span} ${offset} relative`}
		>
			<p className="eyebrow mb-3">{feature.eyebrow}</p>

			{feature.size === 'lead' ? (
				<>
					<h3 className="text-display-md text-foreground max-w-[20ch]">
						{feature.title}
					</h3>
					<p className="drop-cap mt-8 max-w-[58ch] text-base leading-relaxed text-foreground/80 md:text-lg">
						{feature.body}
					</p>
				</>
			) : feature.size === 'medium' ? (
				<>
					<h3 className="font-serif text-2xl leading-tight text-foreground md:text-[1.75rem]">
						{feature.title}
					</h3>
					<p className="mt-4 text-[0.95rem] leading-relaxed text-foreground/70">
						{feature.body}
					</p>
				</>
			) : (
				<>
					<h3 className="font-serif text-xl italic leading-tight text-foreground">
						{feature.title}
					</h3>
					<p className="mt-3 text-sm leading-relaxed text-foreground/65">
						{feature.body}
					</p>
				</>
			)}
		</motion.article>
	);
}

export default FeaturesSection;
