'use client';

import { useRef, useEffect } from 'react';
import { motion } from 'framer-motion';
import { gsap } from 'gsap';
import { ScrollTrigger } from 'gsap/ScrollTrigger';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
	ArrowRight,
	CheckCircle,
	Sparkles,
	Heart
} from 'lucide-react';

gsap.registerPlugin(ScrollTrigger);

const CallToActionSection = () => {
	const sectionRef = useRef(null);

	const benefits = [
		'Free to join and start exchanging',
		'Verified skill assessments & profiles',
		'Community support & guidance',
		'Build meaningful connections'
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
							<div className="flex items-center gap-4 mt-10">
								<div className="flex -space-x-2">
									{[1, 2, 3, 4].map((i) => (
										<div
											key={i}
											className="w-10 h-10 rounded-full border-2 border-sage-900 bg-gradient-to-br from-sage-300 to-terracotta-300 flex items-center justify-center text-stone-700 font-bold text-xs"
										>
											{String.fromCharCode(64 + i)}
										</div>
									))}
								</div>
								<p className="text-stone-300 text-sm">
									Join a growing community of learners and teachers
								</p>
							</div>
						</motion.div>

						{/* Right Content - CTA Card */}
						<motion.div
							initial={{ opacity: 0, x: 50 }}
							whileInView={{ opacity: 1, x: 0 }}
							transition={{ duration: 0.8, delay: 0.2 }}
							viewport={{ once: true }}
							className="floating-form"
						>
							<Card className="bg-white/95 backdrop-blur-lg border-0 shadow-2xl rounded-3xl overflow-hidden">
								<CardContent className="p-10">
									<div className="text-center mb-8">
										<div className="inline-flex p-3 bg-gradient-to-r from-sage-100 to-terracotta-100 rounded-2xl mb-4">
											<Heart className="h-6 w-6 text-sage-600" />
										</div>
										<h3 className="text-3xl font-bold text-stone-900 mb-3">
											Start Your Journey
										</h3>
										<p className="text-stone-600 text-lg leading-relaxed">
											Create your free account and get matched
											with skill exchange partners in your area
										</p>
									</div>

									<div className="space-y-4">
										<a href="/auth/signup">
											<Button
												className="w-full h-14 bg-gradient-to-r from-sage-600 to-terracotta-600 hover:from-sage-700 hover:to-terracotta-700 text-white font-semibold text-lg rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300"
											>
												<div className="flex items-center text-black hover:text-white">
													Create Free Account
													<ArrowRight className="ml-3 h-5 w-5" />
												</div>
											</Button>
										</a>
										<a href="/auth/signin">
											<Button
												variant="outline"
												className="w-full h-12 rounded-2xl text-base mt-3"
											>
												Already have an account? Sign In
											</Button>
										</a>
									</div>

									<div className="mt-8 flex items-center justify-center gap-6 text-sm text-stone-500">
										<div className="flex items-center gap-1">
											<CheckCircle className="h-4 w-4 text-emerald-500" />
											Free forever
										</div>
										<div className="flex items-center gap-1">
											<CheckCircle className="h-4 w-4 text-emerald-500" />
											No credit card
										</div>
									</div>
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
