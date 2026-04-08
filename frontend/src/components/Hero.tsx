'use client';

import { motion } from 'framer-motion';
import { Button } from '@/components/ui/button';
import { ArrowRight } from 'lucide-react';
import Link from 'next/link';

const HeroSection = () => {
	return (
		<section className="relative flex items-center justify-center overflow-hidden bg-background">
			{/* Subtle grid background */}
			<div className="absolute inset-0 bg-[linear-gradient(to_right,var(--border)_1px,transparent_1px),linear-gradient(to_bottom,var(--border)_1px,transparent_1px)] bg-[size:4rem_4rem] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_0%,#000_70%,transparent_110%)] opacity-40" />

			<div className="relative z-10 max-w-4xl mx-auto px-4 pt-32 pb-24 text-center">
				{/* Pill badge */}
				<motion.div
					initial={{ opacity: 0, y: 16 }}
					animate={{ opacity: 1, y: 0 }}
					transition={{ duration: 0.5 }}
					className="inline-flex items-center gap-2 border border-border bg-card px-4 py-1.5 rounded-full mb-8 text-sm text-muted-foreground"
				>
					<span className="w-1.5 h-1.5 rounded-full bg-primary" />
					Peer-to-peer skill exchange platform
				</motion.div>

				{/* Headline */}
				<motion.h1
					initial={{ opacity: 0, y: 20 }}
					animate={{ opacity: 1, y: 0 }}
					transition={{ duration: 0.6, delay: 0.1 }}
					className="text-display-xl text-foreground mb-6"
				>
					Learn anything.
					<br />
					<span className="text-muted-foreground">Teach everything.</span>
				</motion.h1>

				{/* Subheading */}
				<motion.p
					initial={{ opacity: 0, y: 20 }}
					animate={{ opacity: 1, y: 0 }}
					transition={{ duration: 0.6, delay: 0.2 }}
					className="text-lg md:text-xl text-muted-foreground max-w-2xl mx-auto mb-10 leading-relaxed"
				>
					Exchange your expertise directly with others. No middlemen, no fees —
					just real people trading real skills.
				</motion.p>

				{/* CTAs */}
				<motion.div
					initial={{ opacity: 0, y: 20 }}
					animate={{ opacity: 1, y: 0 }}
					transition={{ duration: 0.6, delay: 0.3 }}
					className="flex flex-col sm:flex-row gap-3 justify-center items-center mb-20"
				>
					<Link href="/auth/signup">
						<Button size="lg" className="px-8 h-11 text-sm font-medium group">
							Get Started
							<ArrowRight className="ml-2 h-4 w-4 group-hover:translate-x-0.5 transition-transform" />
						</Button>
					</Link>
					<Link href="/browse">
						<Button variant="outline" size="lg" className="px-8 h-11 text-sm font-medium">
							Browse Skills
						</Button>
					</Link>
				</motion.div>

				{/* Stats */}
				<motion.div
					initial={{ opacity: 0 }}
					animate={{ opacity: 1 }}
					transition={{ duration: 0.8, delay: 0.5 }}
					className="flex items-center justify-center gap-12 md:gap-16"
				>
					{[
						{ value: '15K+', label: 'Active users' },
						{ value: '800+', label: 'Skills listed' },
						{ value: '98%', label: 'Satisfaction' }
					].map((stat) => (
						<div key={stat.label} className="text-center">
							<div className="text-2xl md:text-3xl font-bold text-foreground tracking-tight">
								{stat.value}
							</div>
							<div className="text-xs text-muted-foreground mt-1">
								{stat.label}
							</div>
						</div>
					))}
				</motion.div>
			</div>
		</section>
	);
};

export default HeroSection;
