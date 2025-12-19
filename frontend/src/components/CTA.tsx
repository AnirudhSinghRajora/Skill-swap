'use client';

import type React from 'react';

import { useState, useRef, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { gsap } from 'gsap';
import { ScrollTrigger } from 'gsap/ScrollTrigger';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import {
	ArrowRight,
	CheckCircle,
	Users,
	Star,
	Globe,
	Sparkles,
	Mail,
	Lock,
	Heart,
	Award
} from 'lucide-react';

gsap.registerPlugin(ScrollTrigger);

const CallToActionSection = () => {
	const sectionRef = useRef(null);
	const [email, setEmail] = useState('');
	const [isSubmitted, setIsSubmitted] = useState(false);
	const [isLoading, setIsLoading] = useState(false);

	const socialProofStats = [
		{ icon: Users, value: '15,000+', label: 'Active Members' },
		{ icon: Star, value: '4.9/5', label: 'User Rating' },
		{ icon: Globe, value: '65+', label: 'Countries' }
	];

	const benefits = [
		'Free to join and start exchanging',
		'Verified skill assessments & profiles',
		'24/7 community support & guidance',
		'Earn certificates and digital badges'
	];

	useEffect(() => {
		const ctx = gsap.context(() => {
			// Enhanced floating animation
			gsap.to('.floating-form', {
				y: -15,
				rotation: 1,
				duration: 4,
				repeat: -1,
				yoyo: true,
				ease: 'power2.inOut'
			});

			// Parallax background elements
			gsap.to('.cta-bg-element', {
				y: -80,
				rotation: 10,
				ease: 'none',
				scrollTrigger: {
					trigger: sectionRef.current,
					start: 'top bottom',
					end: 'bottom top',
					scrub: true
				}
			});
		}, sectionRef);

		return () => ctx.revert();
	}, []);

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault();
		setIsLoading(true);

		// Simulate API call
		await new Promise((resolve) => setTimeout(resolve, 2000));

		setIsSubmitted(true);
		setIsLoading(false);
	};

	return (
		<section
			ref={sectionRef}
			className="relative py-24 bg-gradient-to-br from-sage-900 via-terracotta-900 to-amber-900 overflow-hidden"
		>
			{/* Enhanced Background Elements */}
			<div className="absolute inset-0 pointer-events-none">
				<div className="cta-bg-element absolute top-20 left-10 w-40 h-40 bg-sage-400 rounded-full opacity-10 blur-3xl"></div>
				<div className="cta-bg-element absolute top-40 right-20 w-32 h-32 bg-terracotta-400 rounded-full opacity-10 blur-2xl"></div>
				<div className="cta-bg-element absolute bottom-20 left-1/3 w-36 h-36 bg-amber-400 rounded-full opacity-10 blur-3xl"></div>
				<div className="cta-bg-element absolute top-1/2 right-1/4 w-24 h-24 bg-emerald-400 rounded-full opacity-10 blur-xl"></div>
			</div>

			<div className="container mx-auto px-4 relative z-10">
				<div className="max-w-7xl mx-auto">
					<div className="grid lg:grid-cols-2 gap-16 items-center">
						{/* Left Content */}
						<motion.div
							initial={{ opacity: 0, x: -50 }}
							whileInView={{ opacity: 1, x: 0 }}
							transition={{ duration: 0.8 }}
							viewport={{ once: true }}
							className="text-white"
						>
							<div className="mb-8">
								<Badge className="bg-white/15 text-white border-white/20 mb-6 px-4 py-2 rounded-full">
									<Sparkles className="mr-2 h-4 w-4" />
									Join Our Community
								</Badge>
								<h2 className="text-4xl md:text-6xl font-bold mb-8 leading-tight">
									Ready to{' '}
									<span className="bg-gradient-to-r from-sage-300 to-terracotta-300 bg-clip-text text-transparent">
										Transform
									</span>
									<br />
									Your Future?
								</h2>
								<p className="text-xl text-stone-200 mb-10 leading-relaxed">
									Join thousands of professionals who are already growing their
									skills and expanding their networks through meaningful
									exchanges. Your next opportunity awaits.
								</p>
							</div>

							{/* Enhanced Benefits List */}
							<div className="space-y-5 mb-10">
								{benefits.map((benefit, index) => (
									<motion.div
										key={index}
										initial={{ opacity: 0, x: -20 }}
										whileInView={{ opacity: 1, x: 0 }}
										transition={{ duration: 0.5, delay: index * 0.1 }}
										viewport={{ once: true }}
										className="flex items-center group"
									>
										<div className="p-1 bg-emerald-500 rounded-full mr-4 group-hover:scale-110 transition-transform duration-300">
											<CheckCircle className="h-5 w-5 text-white" />
										</div>
										<span className="text-stone-200 text-lg">{benefit}</span>
									</motion.div>
								))}
							</div>

							{/* Enhanced Social Proof */}
							<div className="grid grid-cols-3 gap-8">
								{socialProofStats.map((stat, index) => (
									<motion.div
										key={index}
										initial={{ opacity: 0, y: 20 }}
										whileInView={{ opacity: 1, y: 0 }}
										transition={{ duration: 0.5, delay: 0.3 + index * 0.1 }}
										viewport={{ once: true }}
										className="text-center group"
									>
										<div className="inline-flex p-3 bg-white/15 backdrop-blur-sm rounded-2xl mb-3 group-hover:scale-110 transition-transform duration-300">
											<stat.icon className="h-6 w-6 text-sage-300" />
										</div>
										<div className="text-3xl font-bold text-white mb-1">
											{stat.value}
										</div>
										<div className="text-sm text-stone-300">{stat.label}</div>
									</motion.div>
								))}
							</div>
						</motion.div>

						{/* Right Content - Enhanced Floating Form */}
						<motion.div
							initial={{ opacity: 0, x: 50 }}
							whileInView={{ opacity: 1, x: 0 }}
							transition={{ duration: 0.8, delay: 0.2 }}
							viewport={{ once: true }}
							className="floating-form"
						>
							<Card className="bg-white/95 backdrop-blur-lg border-0 shadow-2xl rounded-3xl overflow-hidden">
								<CardContent className="p-10">
									<AnimatePresence mode="wait">
										{!isSubmitted ? (
											<motion.div
												key="form"
												initial={{ opacity: 0 }}
												animate={{ opacity: 1 }}
												exit={{ opacity: 0 }}
												transition={{ duration: 0.3 }}
											>
												<div className="text-center mb-8">
													<div className="inline-flex p-3 bg-gradient-to-r from-sage-100 to-terracotta-100 rounded-2xl mb-4">
														<Heart className="h-6 w-6 text-sage-600" />
													</div>
													<h3 className="text-3xl font-bold text-stone-900 mb-3">
														Start Your Journey
													</h3>
													<p className="text-stone-600 text-lg leading-relaxed">
														Create your free account and get matched
														with your first skill exchange partner
														within 24 hours
													</p>
												</div>

												<form onSubmit={handleSubmit} className="space-y-6">
													<div className="relative">
														<Mail className="absolute left-4 top-1/2 transform -translate-y-1/2 h-5 w-5 text-stone-400" />
														<Input
															type="email"
															placeholder="Enter your email address"
															value={email}
															onChange={(e) =>
																setEmail(e.target.value)
															}
															required
															className="pl-12 h-14 text-lg rounded-2xl border-2 border-stone-200 focus:border-sage-400"
														/>
													</div>

													<Button
														type="submit"
														disabled={isLoading}
														className="w-full h-14 bg-gradient-to-r from-sage-600 to-terracotta-600 hover:from-sage-700 hover:to-terracotta-700 text-white font-semibold text-lg rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300"
													>
														{isLoading ? (
															<div className="flex items-center">
																<div className="animate-spin rounded-full h-5 w-5 border-b-2 border-white mr-3"></div>
																Creating Account...
															</div>
														) : (
															<div className="flex items-center text-black hover:text-white">
																Get Started Free
																<ArrowRight className="ml-3 h-5 w-5" />
															</div>
														)}
													</Button>
												</form>

												<div className="mt-8 text-center">
													<div className="flex items-center justify-center text-sm text-stone-500 mb-3">
														<Lock className="h-4 w-4 mr-2" />
														100% secure and private
													</div>
													<p className="text-xs text-stone-400 leading-relaxed">
														By signing up, you agree to our Terms of
														Service and Privacy Policy. No spam,
														unsubscribe anytime.
													</p>
												</div>
											</motion.div>
										) : (
											<motion.div
												key="success"
												initial={{ opacity: 0, scale: 0.8 }}
												animate={{ opacity: 1, scale: 1 }}
												exit={{ opacity: 0, scale: 0.8 }}
												transition={{ duration: 0.5 }}
												className="text-center py-10"
											>
												<motion.div
													initial={{ scale: 0 }}
													animate={{ scale: 1 }}
													transition={{ duration: 0.5, delay: 0.2 }}
													className="w-20 h-20 bg-gradient-to-r from-emerald-500 to-teal-500 rounded-full flex items-center justify-center mx-auto mb-6"
												>
													<CheckCircle className="h-10 w-10 text-white" />
												</motion.div>

												<h3 className="text-3xl font-bold text-stone-900 mb-4">
													Welcome to SkillShare Connect!
												</h3>
												<p className="text-stone-600 mb-8 text-lg leading-relaxed">
													Check your email to complete your account setup
													and start connecting with skill exchange
													partners in your area.
												</p>

												<div className="flex items-center justify-center gap-2 mb-6">
													<Award className="h-5 w-5 text-amber-500" />
													<span className="text-stone-600">
														You&apos;ve earned your first badge: Community
														Member
													</span>
												</div>

												<Button className="bg-gradient-to-r from-sage-600 to-terracotta-600 hover:from-sage-700 hover:to-terracotta-700 text-white px-10 py-4 rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300">
													Go to Dashboard
												</Button>
											</motion.div>
										)}
									</AnimatePresence>
								</CardContent>
							</Card>
						</motion.div>
					</div>
				</div>
			</div>
		</section>
	);
};

export default CallToActionSection;
