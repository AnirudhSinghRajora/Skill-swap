'use client';

import { useRef } from 'react';
import { motion, useInView } from 'framer-motion';
import { Button } from '@/components/ui/button';
import { UserPlus, Search, MessageCircle, TrendingUp, ArrowRight } from 'lucide-react';
import Link from 'next/link';

const HowItWorksSection = () => {
	const ref = useRef(null);
	const inView = useInView(ref, { once: true, amount: 0.1 });

	const steps = [
		{
			icon: UserPlus,
			title: 'Create your profile',
			description: 'Showcase your expertise and define what you want to learn. Set your availability and skill levels.',
		},
		{
			icon: Search,
			title: 'Find a match',
			description: 'Browse profiles and connect with compatible partners who have the skills you need.',
		},
		{
			icon: MessageCircle,
			title: 'Start exchanging',
			description: 'Schedule sessions, chat directly, and begin teaching and learning with your partner.',
		},
		{
			icon: TrendingUp,
			title: 'Grow together',
			description: 'Track your progress, leave reviews, and build lasting professional relationships.',
		},
	];

	return (
		<section className="py-24 md:py-32 border-t border-border">
			<div ref={ref} className="max-w-6xl mx-auto px-4">
				{/* Header */}
				<motion.div
					initial={{ opacity: 0, y: 20 }}
					animate={inView ? { opacity: 1, y: 0 } : {}}
					transition={{ duration: 0.6 }}
					className="max-w-2xl mb-16"
				>
					<p className="text-sm font-medium text-primary mb-3 tracking-wide uppercase">How it works</p>
					<h2 className="text-display-lg text-foreground mb-4">
						Four steps to start learning
					</h2>
					<p className="text-lg text-muted-foreground leading-relaxed">
						Getting started takes less than five minutes.
					</p>
				</motion.div>

				{/* Steps */}
				<div className="grid md:grid-cols-2 lg:grid-cols-4 gap-px bg-border rounded-xl overflow-hidden mb-20">
					{steps.map((step, index) => (
						<motion.div
							key={step.title}
							initial={{ opacity: 0, y: 20 }}
							animate={inView ? { opacity: 1, y: 0 } : {}}
							transition={{ duration: 0.5, delay: 0.1 * index }}
							className="bg-card p-8 relative"
						>
							<span className="text-xs font-mono text-muted-foreground mb-6 block">
								0{index + 1}
							</span>
							<step.icon className="h-5 w-5 text-foreground mb-4" strokeWidth={1.5} />
							<h3 className="font-medium text-foreground mb-2">{step.title}</h3>
							<p className="text-sm text-muted-foreground leading-relaxed">
								{step.description}
							</p>
						</motion.div>
					))}
				</div>

				{/* CTA */}
				<motion.div
					initial={{ opacity: 0, y: 20 }}
					animate={inView ? { opacity: 1, y: 0 } : {}}
					transition={{ duration: 0.6, delay: 0.4 }}
					className="text-center"
				>
					<p className="text-muted-foreground mb-6">
						Join thousands already exchanging skills on SkillSwap.
					</p>
					<div className="flex flex-col sm:flex-row gap-3 justify-center">
						<Button asChild>
							<Link href="/auth/signup">
								Get started free
								<ArrowRight className="ml-2 h-4 w-4" />
							</Link>
						</Button>
						<Button variant="outline" asChild>
							<Link href="/browse">Browse skills</Link>
						</Button>
					</div>
				</motion.div>
			</div>
		</section>
	);
};

export default HowItWorksSection;
