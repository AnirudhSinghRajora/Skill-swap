'use client';

import { useEffect, useRef } from 'react';
import { motion } from 'framer-motion';
import { gsap } from 'gsap';
import { ScrollTrigger } from 'gsap/ScrollTrigger';
import { useInView } from 'react-intersection-observer';
import { Card, CardContent } from '@/components/ui/card';
import { Users, Shield, Zap, Target, Star, Award, Heart, Globe } from 'lucide-react';

gsap.registerPlugin(ScrollTrigger);

const FeaturesSection = () => {
	const sectionRef = useRef(null);
	const [ref, inView] = useInView({
		threshold: 0.1,
		triggerOnce: true
	});

	useEffect(() => {
		const ctx = gsap.context(() => {
			const cards = gsap.utils.toArray('.feature-card') as Element[];

			cards.forEach((card, index) => {
				gsap.fromTo(
					card,
					{ y: 80, opacity: 0, rotateY: 15 },
					{
						y: 0,
						opacity: 1,
						rotateY: 0,
						duration: 0.8,
						delay: index * 0.15,
						scrollTrigger: {
							trigger: card,
							start: 'top 85%',
							end: 'bottom 20%',
							toggleActions: 'play none none reverse'
						}
					}
				);
			});
		}, sectionRef);

		return () => ctx.revert();
	}, []);

	const features = [
		{
			icon: Users,
			title: 'Global Community',
			description:
				'Connect with passionate learners and experts from every corner of the world.',
			stats: '50+ Countries',
			color: 'from-sage-500 to-sage-600'
		},
		{
			icon: Shield,
			title: 'Trust & Safety',
			description: 'Comprehensive verification and review system ensures quality exchanges.',
			stats: '99% Verified',
			color: 'from-terracotta-500 to-terracotta-600'
		},
		{
			icon: Zap,
			title: 'Smart Matching',
			description:
				'AI-powered algorithm finds your perfect skill exchange partner instantly.',
			stats: '< 1 minute',
			color: 'from-amber-500 to-amber-600'
		},
		{
			icon: Target,
			title: 'Goal Achievement',
			description: 'Set learning milestones and track your progress with detailed insights.',
			stats: '95% Complete',
			color: 'from-emerald-500 to-emerald-600'
		},
		{
			icon: Star,
			title: 'Quality Assured',
			description: 'Community-driven ratings ensure exceptional learning experiences.',
			stats: '4.9/5 Rating',
			color: 'from-violet-500 to-violet-600'
		},
		{
			icon: Award,
			title: 'Recognition',
			description: 'Earn verified certificates and showcase your newly acquired skills.',
			stats: '1000+ Badges',
			color: 'from-rose-500 to-rose-600'
		}
	];

	const testimonials = [
		{
			name: 'Maya Patel',
			role: 'Product Designer',
			text: 'Traded my design skills for data science knowledge. The community here is incredible!',
			avatar: '/placeholder.svg?height=60&width=60',
			skill: 'UI/UX → Python'
		},
		{
			name: 'James Chen',
			role: 'Software Engineer',
			text: 'Found my photography mentor in just one day. The matching system is phenomenal.',
			avatar: '/placeholder.svg?height=60&width=60',
			skill: 'React → Photography'
		},
		{
			name: 'Sofia Rodriguez',
			role: 'Marketing Specialist',
			text: 'The verification process gave me confidence. Every exchange has been valuable.',
			avatar: '/placeholder.svg?height=60&width=60',
			skill: 'Marketing → French'
		}
	];

	return (
		<section
			ref={sectionRef}
			className="relative py-24 dark:from-sage-900 dark:via-terracotta-900 dark:to-amber-900 overflow-hidden"
		>
			{/* Organic Background Elements */}
			<div className="absolute inset-0 pointer-events-none">
				<div className="absolute top-20 left-10 w-72 h-72 bg-gradient-to-br from-sage-200/30 to-terracotta-200/30 rounded-full blur-3xl"></div>
				<div className="absolute bottom-20 right-10 w-64 h-64 bg-gradient-to-br from-amber-200/30 to-emerald-200/30 rounded-full blur-3xl"></div>
			</div>

			<div ref={ref} className="container mx-auto px-4 relative z-10">
				{/* Section Header */}
				<motion.div
					initial={{ opacity: 0, y: 50 }}
					animate={inView ? { opacity: 1, y: 0 } : {}}
					transition={{ duration: 0.8 }}
					className="text-center mb-20"
				>
					<div className="inline-flex items-center gap-2 bg-white/60 backdrop-blur-sm px-6 py-3 rounded-full mb-6">
						<Heart className="h-5 w-5 text-terracotta-600" />
						<span className="text-sage-700 dark:text-sage-200 font-medium">
							Why Choose Us
						</span>
					</div>

					<h2 className="text-4xl md:text-6xl font-bold mb-8 text-stone-900 dark:text-stone-100">
						Built for{' '}
						<span className="bg-amber-500 bg-clip-text text-transparent">
							Real Learning
						</span>
					</h2>
					<p className="text-xl text-stone-600 dark:text-stone-300 max-w-3xl mx-auto leading-relaxed">
						Every feature is designed to create meaningful connections and accelerate
						your professional growth
					</p>
				</motion.div>

				{/* Features Grid */}
				<div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8 mb-24">
					{features.map((feature, index) => (
						<motion.div
							key={index}
							className="feature-card"
							whileHover={{ y: -8, scale: 1.02 }}
							transition={{ duration: 0.3 }}
						>
							<Card className="h-full border-0 shadow-lg hover:shadow-2xl transition-all duration-500 bg-white/80 backdrop-blur-sm dark:bg-stone-800/80 rounded-3xl overflow-hidden group">
								<CardContent className="p-8">
									<div className="flex items-start justify-between mb-6">
										<div
											className={`p-4 bg-gradient-to-r ${feature.color} rounded-2xl group-hover:scale-110 transition-transform duration-300`}
										>
											<feature.icon className="h-7 w-7 text-white" />
										</div>
										<div className="text-right">
											<div className="text-sm font-bold text-stone-500 dark:text-stone-400">
												{feature.stats}
											</div>
										</div>
									</div>
									<h3 className="text-2xl font-bold mb-4 text-stone-900 dark:text-stone-100">
										{feature.title}
									</h3>
									<p className="text-stone-600 dark:text-stone-300 leading-relaxed">
										{feature.description}
									</p>
								</CardContent>
							</Card>
						</motion.div>
					))}
				</div>

				{/* Testimonials */}
				<motion.div
					initial={{ opacity: 0, y: 50 }}
					animate={inView ? { opacity: 1, y: 0 } : {}}
					transition={{ duration: 0.8, delay: 0.5 }}
					className="text-center"
				>
					<div className="inline-flex items-center gap-2 bg-white/60 backdrop-blur-sm px-6 py-3 rounded-full mb-8">
						<Globe className="h-5 w-5 text-sage-600" />
						<span className="text-sage-700 dark:text-sage-200 font-medium">
							Success Stories
						</span>
					</div>

					<h3 className="text-3xl md:text-4xl font-bold mb-16 text-stone-900 dark:text-stone-100">
						Real People, Real Growth
					</h3>

					<div className="grid md:grid-cols-3 gap-8">
						{testimonials.map((testimonial, index) => (
							<motion.div
								key={index}
								initial={{ opacity: 0, scale: 0.9 }}
								animate={inView ? { opacity: 1, scale: 1 } : {}}
								transition={{ duration: 0.5, delay: 0.7 + index * 0.2 }}
								whileHover={{ scale: 1.05, y: -5 }}
							>
								<Card className="p-8 border-0 shadow-lg hover:shadow-2xl transition-all duration-500 bg-white/80 backdrop-blur-sm dark:bg-stone-800/80 rounded-3xl">
									<CardContent className="p-0 text-center">
										<img
											src={testimonial.avatar || '/placeholder.svg'}
											alt={testimonial.name}
											className="w-16 h-16 rounded-full mx-auto mb-4 object-cover border-4 border-white shadow-lg"
										/>
										<div className="inline-block bg-gradient-to-r from-sage-100 to-terracotta-100 dark:from-sage-800 dark:to-terracotta-800 px-4 py-2 rounded-full text-sm font-medium text-stone-700 dark:text-stone-300 mb-4">
											{testimonial.skill}
										</div>
										<p className="text-stone-600 dark:text-stone-300 mb-6 italic leading-relaxed">
										&quot;{testimonial.text}&quot;
										</p>
										<div>
											<div className="font-bold text-stone-900 dark:text-stone-100 text-lg">
												{testimonial.name}
											</div>
											<div className="text-stone-500 dark:text-stone-400">
												{testimonial.role}
											</div>
										</div>
									</CardContent>
								</Card>
							</motion.div>
						))}
					</div>
				</motion.div>
			</div>
		</section>
	);
};

export default FeaturesSection;
