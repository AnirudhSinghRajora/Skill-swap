'use client';

import { useEffect, useRef, useState, useMemo } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { gsap } from 'gsap';
import { ScrollTrigger } from 'gsap/ScrollTrigger';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { UserPlus, Search, MessageCircle, TrendingUp, Play, CheckCircle } from 'lucide-react';

gsap.registerPlugin(ScrollTrigger);

const HowItWorksSection = () => {
	const sectionRef = useRef(null);
	const [activeStep, setActiveStep] = useState(0);
	const [isPlaying, setIsPlaying] = useState(false);

	const steps = useMemo(() => [
		{
			icon: UserPlus,
			title: 'Create Your Profile',
			description:
				'Showcase your expertise and learning goals with a compelling profile that attracts the right connections.',
			details:
				'Upload your portfolio, set availability, define skill levels, and let our community know what makes you unique.',
			color: 'from-sage-500 to-emerald-500'
		},
		{
			icon: Search,
			title: 'Discover Matches',
			description:
				'Our intelligent algorithm connects you with compatible skill exchange partners based on your goals.',
			details:
				'Browse verified profiles, check compatibility scores, and use advanced filters to find your perfect learning partner.',
			color: 'from-terracotta-500 to-rose-500'
		},
		{
			icon: MessageCircle,
			title: 'Start Learning',
			description:
				'Connect with matches, schedule sessions, and begin your transformative skill exchange journey.',
			details:
				'Use integrated video chat, collaborative tools, and structured learning paths to maximize your growth.',
			color: 'from-amber-500 to-orange-500'
		},
		{
			icon: TrendingUp,
			title: 'Track Growth',
			description:
				'Monitor progress, earn recognition, and build lasting professional relationships as you advance.',
			details:
				'Set milestones, receive feedback, showcase achievements, and expand your network for future opportunities.',
			color: 'from-violet-500 to-purple-500'
		}
	], []);

	useEffect(() => {
		const ctx = gsap.context(() => {
			// Animate step cards with more sophisticated entrance
			steps.forEach((_, index) => {
				gsap.fromTo(
					`.step-card-${index}`,
					{ x: index % 2 === 0 ? -120 : 120, opacity: 0, rotateY: 15 },
					{
						x: 0,
						opacity: 1,
						rotateY: 0,
						duration: 1,
						ease: 'power3.out',
						scrollTrigger: {
							trigger: `.step-card-${index}`,
							start: 'top 85%',
							end: 'bottom 20%',
							toggleActions: 'play none none reverse'
						}
					}
				);
			});

			// Enhanced connecting lines animation
			gsap.fromTo(
				'.connecting-line',
				{ scaleY: 0, opacity: 0 },
				{
					scaleY: 1,
					opacity: 1,
					duration: 1.2,
					stagger: 0.4,
					ease: 'power2.out',
					scrollTrigger: {
						trigger: '.steps-container',
						start: 'top 60%',
						end: 'bottom 40%',
						toggleActions: 'play none none reverse'
					}
				}
			);
		}, sectionRef);

		return () => ctx.revert();
	}, [steps]);

	const handlePlayDemo = () => {
		setIsPlaying(true);
		let currentStep = 0;
		const interval = setInterval(() => {
			currentStep = (currentStep + 1) % steps.length;
			setActiveStep(currentStep);

			if (currentStep === 0) {
				clearInterval(interval);
				setIsPlaying(false);
			}
		}, 2500);
	};

	return (
		<section
			ref={sectionRef}
			className="py-24 dark:from-stone-900 dark:via-terracotta-900 dark:to-sage-900"
		>
			<div className="container mx-auto px-4">
				{/* Section Header */}
				<motion.div
					initial={{ opacity: 0, y: 50 }}
					whileInView={{ opacity: 1, y: 0 }}
					transition={{ duration: 0.8 }}
					viewport={{ once: true }}
					className="text-center mb-20"
				>
					<div className="inline-flex items-center gap-2 bg-white/60 backdrop-blur-sm px-6 py-3 rounded-full mb-6">
						<CheckCircle className="h-5 w-5 text-emerald-600" />
						<span className="text-stone-700 dark:text-stone-300 font-medium">
							Simple Process
						</span>
					</div>

					<h2 className="text-4xl md:text-6xl font-bold mb-8 text-stone-900 dark:text-stone-100">
						How It{' '}
						<span className="bg-amber-500 bg-clip-text text-transparent">Works</span>
					</h2>
					<p className="text-xl text-stone-600 dark:text-stone-300 max-w-3xl mx-auto mb-10 leading-relaxed">
						Start your skill exchange journey in four simple steps. Join thousands who
						are already growing their careers.
					</p>

					<Button
						onClick={handlePlayDemo}
						disabled={isPlaying}
						className="text-black bg-gradient-to-r from-sage-600 to-terracotta-600 hover:from-sage-700 hover:to-terracotta-700 hover:text-white px-8 py-4 rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300"
					>
						<Play className="mr-3 h-5 w-5" />
						{isPlaying ? 'Playing Demo...' : 'Watch Interactive Demo'}
					</Button>
				</motion.div>

				{/* Steps Container */}
				<div className="steps-container relative max-w-5xl mx-auto">
					{steps.map((step, index) => (
						<div key={index} className="relative mb-20 last:mb-0">
							{/* Enhanced Connecting Line */}
							{index < steps.length - 1 && (
								<div className="absolute left-1/2 top-full w-1 h-20 transform -translate-x-1/2 z-0">
									<div
										className="connecting-line w-full h-full bg-gradient-to-b from-sage-400 to-terracotta-400 rounded-full"
										style={{ transformOrigin: 'top' }}
									/>
								</div>
							)}

							{/* Step Card */}
							<motion.div
								className={`step-card-${index} relative z-10`}
								onMouseEnter={() => setActiveStep(index)}
								whileHover={{ scale: 1.02, y: -5 }}
								transition={{ duration: 0.3 }}
							>
								<Card
									className={`transition-all duration-500 rounded-3xl overflow-hidden ${
										activeStep === index
											? 'shadow-2xl border-2 border-sage-300 dark:border-sage-600 bg-white dark:bg-stone-800'
											: 'shadow-lg hover:shadow-xl bg-white/80 backdrop-blur-sm dark:bg-stone-800/80'
									}`}
								>
									<CardContent className="p-10">
										<div className="flex flex-col lg:flex-row items-center text-center lg:text-left">
											{/* Step Icon */}
											<div
												className={`flex-shrink-0 mb-8 lg:mb-0 lg:mr-10 p-6 rounded-3xl transition-all duration-500 ${
													activeStep === index
														? `bg-gradient-to-r ${step.color} scale-110 shadow-lg`
														: 'bg-stone-100 dark:bg-stone-700'
												}`}
											>
												<step.icon
													className={`h-10 w-10 ${
														activeStep === index
															? 'text-white'
															: 'text-stone-600 dark:text-stone-300'
													}`}
												/>
											</div>

											{/* Step Content */}
											<div className="flex-1">
												<div className="flex items-center justify-center lg:justify-start mb-6">
													<span className="text-sm font-bold text-sage-600 dark:text-sage-400 mr-4 bg-sage-100 dark:bg-sage-800 px-3 py-1 rounded-full">
														Step {index + 1}
													</span>
													<h3 className="text-2xl lg:text-3xl font-bold text-stone-900 dark:text-stone-100">
														{step.title}
													</h3>
												</div>

												<p className="text-lg text-stone-600 dark:text-stone-300 mb-6 leading-relaxed">
													{step.description}
												</p>

												<AnimatePresence>
													{activeStep === index && (
														<motion.p
															initial={{ opacity: 0, height: 0 }}
															animate={{ opacity: 1, height: 'auto' }}
															exit={{ opacity: 0, height: 0 }}
															transition={{ duration: 0.4 }}
															className="text-stone-500 dark:text-stone-400 leading-relaxed"
														>
															{step.details}
														</motion.p>
													)}
												</AnimatePresence>
											</div>

											{/* Step Number */}
											<div
												className={`flex-shrink-0 mt-8 lg:mt-0 lg:ml-10 w-16 h-16 rounded-2xl flex items-center justify-center text-3xl font-bold transition-all duration-500 ${
													activeStep === index
														? `bg-gradient-to-r ${step.color} text-white scale-110 shadow-lg`
														: 'bg-stone-100 dark:bg-stone-700 text-stone-600 dark:text-stone-300'
												}`}
											>
												{index + 1}
											</div>
										</div>
									</CardContent>
								</Card>
							</motion.div>
						</div>
					))}
				</div>

				{/* Call to Action */}
				<motion.div
					initial={{ opacity: 0, y: 50 }}
					whileInView={{ opacity: 1, y: 0 }}
					transition={{ duration: 0.8, delay: 0.3 }}
					viewport={{ once: true }}
					className="mt-24 text-center"
				>
					<div className="max-w-3xl mx-auto p-10 bg-gradient-to-r from-sage-100 to-terracotta-100 dark:from-sage-800 dark:to-terracotta-800 rounded-3xl shadow-xl">
						<h3 className="text-2xl lg:text-3xl font-bold mb-6 text-stone-900 dark:text-stone-100">
							Ready to Start Your Journey?
						</h3>
						<p className="text-lg text-stone-600 dark:text-stone-300 mb-8">
							Join our community of learners and start exchanging skills today.
						</p>
						<div className="flex flex-col sm:flex-row gap-4 justify-center">
							<Button className="bg-gradient-to-r from-sage-600 to-terracotta-600 hover:from-sage-700 hover:to-terracotta-700 text-black hover:text-white px-10 py-4 rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300">
								Join Now - It&apos;s Free
							</Button>
							<Button
								variant="outline"
								className="px-10 py-4 border-2 rounded-2xl bg-transparent"
							>
								Learn More
							</Button>
						</div>
					</div>
				</motion.div>
			</div>
		</section>
	);
};

export default HowItWorksSection;
