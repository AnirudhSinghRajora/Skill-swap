'use client';

import { motion } from 'framer-motion';
import { useInView } from 'react-intersection-observer';
import { Users, Shield, Zap, Target, Star, Award } from 'lucide-react';

const FeaturesSection = () => {
	const [ref, inView] = useInView({ threshold: 0.1, triggerOnce: true });

	const features = [
		{
			icon: Users,
			title: 'Global Community',
			description: 'Connect with learners and experts across 50+ countries. Find your perfect match regardless of location.'
		},
		{
			icon: Shield,
			title: 'Trust & Safety',
			description: 'Every profile is verified. Our review system ensures quality exchanges and builds real accountability.'
		},
		{
			icon: Zap,
			title: 'Smart Matching',
			description: 'Our algorithm pairs you with the right partner based on skills, goals, and availability — in under a minute.'
		},
		{
			icon: Target,
			title: 'Goal Tracking',
			description: 'Set learning milestones and track progress with detailed insights. Stay accountable to your growth.'
		},
		{
			icon: Star,
			title: 'Quality Assured',
			description: 'Community-driven ratings ensure every exchange is valuable. 4.9 average across thousands of sessions.'
		},
		{
			icon: Award,
			title: 'Recognition',
			description: 'Earn verified credentials and showcase your skills. Build a portfolio that speaks for itself.'
		}
	];

	return (
		<section className="py-24 md:py-32 bg-background">
			<div ref={ref} className="max-w-6xl mx-auto px-4">
				{/* Section Header */}
				<motion.div
					initial={{ opacity: 0, y: 20 }}
					animate={inView ? { opacity: 1, y: 0 } : {}}
					transition={{ duration: 0.6 }}
					className="max-w-2xl mb-16"
				>
					<p className="text-sm font-medium text-primary mb-3 tracking-wide uppercase">Features</p>
					<h2 className="text-display-lg text-foreground mb-4">
						Everything you need to grow
					</h2>
					<p className="text-lg text-muted-foreground leading-relaxed">
						Built with care for people who take learning seriously.
					</p>
				</motion.div>

				{/* Features Grid */}
				<div className="grid md:grid-cols-2 lg:grid-cols-3 gap-x-8 gap-y-12">
					{features.map((feature, index) => (
						<motion.div
							key={index}
							initial={{ opacity: 0, y: 20 }}
							animate={inView ? { opacity: 1, y: 0 } : {}}
							transition={{ duration: 0.5, delay: index * 0.08 }}
						>
							<feature.icon className="h-5 w-5 text-primary mb-4" strokeWidth={1.5} />
							<h3 className="text-base font-semibold text-foreground mb-2">
								{feature.title}
							</h3>
							<p className="text-sm text-muted-foreground leading-relaxed">
								{feature.description}
							</p>
						</motion.div>
					))}
				</div>
			</div>
		</section>
	);
};

export default FeaturesSection;
