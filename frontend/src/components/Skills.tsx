'use client';

import { useEffect, useRef, useState } from 'react';
import { motion } from 'framer-motion';
import { gsap } from 'gsap';
import { ScrollTrigger } from 'gsap/ScrollTrigger';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
	Code,
	Palette,
	BarChart,
	Megaphone,
	Camera,
	Music,
	Languages,
	Wrench,
	Search,
	Filter,
	TrendingUp,
	Sparkles,
	ArrowRight
} from 'lucide-react';

gsap.registerPlugin(ScrollTrigger);

const SkillsMarketplace = () => {
	const sectionRef = useRef(null);
	const [searchTerm, setSearchTerm] = useState('');
	const [selectedCategory, setSelectedCategory] = useState('All');
	const [skillCount, setSkillCount] = useState(0);

	const skillCategories = [
		{
			icon: Code,
			name: 'Programming',
			count: 156,
			color: 'from-sage-500 to-emerald-500',
			bgColor: 'bg-sage-100 dark:bg-sage-800'
		},
		{
			icon: Palette,
			name: 'Design',
			count: 124,
			color: 'from-terracotta-500 to-rose-500',
			bgColor: 'bg-terracotta-100 dark:bg-terracotta-800'
		},
		{
			icon: BarChart,
			name: 'Analytics',
			count: 89,
			color: 'from-emerald-500 to-teal-500',
			bgColor: 'bg-emerald-100 dark:bg-emerald-800'
		},
		{
			icon: Megaphone,
			name: 'Marketing',
			count: 112,
			color: 'from-amber-500 to-orange-500',
			bgColor: 'bg-amber-100 dark:bg-amber-800'
		},
		{
			icon: Camera,
			name: 'Photography',
			count: 67,
			color: 'from-violet-500 to-purple-500',
			bgColor: 'bg-violet-100 dark:bg-violet-800'
		},
		{
			icon: Music,
			name: 'Music',
			count: 45,
			color: 'from-rose-500 to-pink-500',
			bgColor: 'bg-rose-100 dark:bg-rose-800'
		},
		{
			icon: Languages,
			name: 'Languages',
			count: 78,
			color: 'from-indigo-500 to-blue-500',
			bgColor: 'bg-indigo-100 dark:bg-indigo-800'
		},
		{
			icon: Wrench,
			name: 'Engineering',
			count: 134,
			color: 'from-stone-500 to-gray-500',
			bgColor: 'bg-stone-100 dark:bg-stone-800'
		}
	];

	const categories = ['All', ...skillCategories.map((cat) => cat.name)];
	const totalSkills = skillCategories.reduce((sum, cat) => sum + cat.count, 0);

	const filteredCategories =
		selectedCategory === 'All'
			? skillCategories
			: skillCategories.filter((cat) => cat.name === selectedCategory);

	useEffect(() => {
		// Animate skill counter
		gsap.to(
			{ count: 0 },
			{
				count: totalSkills,
				duration: 2.5,
				ease: 'power2.out',
				scrollTrigger: {
					trigger: sectionRef.current,
					start: 'top 80%',
					onEnter: () => {
						gsap.to(
							{ count: 0 },
							{
								count: totalSkills,
								duration: 2.5,
								ease: 'power2.out',
								onUpdate: function () {
									setSkillCount(Math.round(this.targets()[0].count));
								}
							}
						);
					}
				}
			}
		);

		const ctx = gsap.context(() => {
			// Animate category cards with stagger
			gsap.fromTo(
				'.skill-category',
				{ y: 60, opacity: 0, scale: 0.8, rotateY: 15 },
				{
					y: 0,
					opacity: 1,
					scale: 1,
					rotateY: 0,
					duration: 0.8,
					stagger: 0.12,
					scrollTrigger: {
						trigger: '.skills-grid',
						start: 'top 85%',
						end: 'bottom 20%',
						toggleActions: 'play none none reverse'
					}
				}
			);
		}, sectionRef);

		return () => ctx.revert();
	}, [totalSkills]);

	return (
		<section
			ref={sectionRef}
			className="py-24  dark:from-stone-900 dark:via-sage-900 dark:to-amber-900"
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
						<Sparkles className="h-5 w-5 text-amber-600" />
						<span className="text-stone-700 dark:text-stone-300 font-medium">
							Explore Skills
						</span>
					</div>

					<h2 className="text-4xl md:text-6xl font-bold mb-8 text-stone-900 dark:text-stone-100">
						Skills{' '}
						<span className="bg-amber-500 bg-clip-text text-transparent">
							Marketplace
						</span>
					</h2>
					<p className="text-xl text-stone-600 dark:text-stone-300 max-w-3xl mx-auto mb-8 leading-relaxed">
						Discover a world of knowledge. From coding to cooking, find the perfect
						skill exchange partner.
					</p>

					{/* Enhanced Counter */}
					<motion.div
						initial={{ scale: 0 }}
						whileInView={{ scale: 1 }}
						transition={{ duration: 0.8, delay: 0.3 }}
						viewport={{ once: true }}
						className="inline-flex items-center gap-3 bg-gradient-to-r from-sage-600 to-terracotta-600 text-white px-8 py-4 rounded-2xl text-lg font-bold shadow-lg"
					>
						<TrendingUp className="h-6 w-6 text-black" />
						<span className="text-black font-mono text-2xl">{skillCount}</span>
						<span className="text-black">Active Skills</span>
					</motion.div>
				</motion.div>

				{/* Enhanced Search and Filter */}
				<motion.div
					initial={{ opacity: 0, y: 30 }}
					whileInView={{ opacity: 1, y: 0 }}
					transition={{ duration: 0.8, delay: 0.2 }}
					viewport={{ once: true }}
					className="max-w-3xl mx-auto mb-16"
				>
					<div className="flex flex-col sm:flex-row gap-4">
						<div className="relative flex-1">
							<Search className="absolute left-4 top-1/2 transform -translate-y-1/2 h-5 w-5 text-stone-400" />
							<Input
								placeholder="Search for any skill..."
								value={searchTerm}
								onChange={(e) => setSearchTerm(e.target.value)}
								className="pl-12 h-14 text-lg rounded-2xl border-2 border-stone-200 dark:border-stone-700 bg-white/80 backdrop-blur-sm"
							/>
						</div>
						<Button
							variant="outline"
							className="h-14 px-8 border-2 border-stone-200 dark:border-stone-700 hover:bg-stone-50 dark:hover:bg-stone-800 rounded-2xl bg-transparent"
						>
							<Filter className="mr-2 h-5 w-5" />
							Filters
						</Button>
					</div>
				</motion.div>

				{/* Category Filter */}
				<motion.div
					initial={{ opacity: 0, y: 30 }}
					whileInView={{ opacity: 1, y: 0 }}
					transition={{ duration: 0.8, delay: 0.3 }}
					viewport={{ once: true }}
					className="mb-16"
				>
					<div className="flex flex-wrap justify-center gap-3">
						{categories.map((category) => (
							<Button
								key={category}
								variant={selectedCategory === category ? 'default' : 'outline'}
								onClick={() => setSelectedCategory(category)}
								className={`transition-all duration-300 rounded-full px-6 py-3 ${
									selectedCategory === category
										? 'bg-gradient-to-r from-sage-600 to-terracotta-600 text-black hover:text-white shadow-lg scale-105'
										: 'bg-gradient-to-r from-sage-600 to-terracotta-600 text-black hover:text-white shadow-lg scale-105'
								}`}
							>
								{category}
							</Button>
						))}
					</div>
				</motion.div>

				{/* Enhanced Skills Grid */}
				<div className="skills-grid grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-8 mb-20">
					{filteredCategories.map((skill) => (
						<motion.div
							key={skill.name}
							className="skill-category"
							whileHover={{ y: -12, scale: 1.05 }}
							transition={{ duration: 0.3 }}
						>
							<Card className="h-full border-0 shadow-lg hover:shadow-2xl transition-all duration-500 cursor-pointer group bg-white/80 backdrop-blur-sm dark:bg-stone-800/80 rounded-3xl overflow-hidden">
								<CardContent className="p-8 text-center">
									<div
										className={`inline-flex p-5 rounded-3xl ${skill.bgColor} mb-6 group-hover:scale-110 transition-transform duration-300`}
									>
										<skill.icon className="h-8 w-8 text-stone-700 dark:text-stone-300" />
									</div>

									<h3 className="text-xl font-bold mb-3 text-stone-900 dark:text-stone-100">
										{skill.name}
									</h3>

									<p className="text-stone-600 dark:text-stone-300 mb-4">
										{skill.count} skilled professionals
									</p>

									<div className="flex items-center justify-center gap-2">
										<Badge
											variant="secondary"
											className="bg-stone-100 dark:bg-stone-700 text-stone-600 dark:text-stone-300 px-3 py-1"
										>
											High Demand
										</Badge>
									</div>
								</CardContent>
							</Card>
						</motion.div>
					))}
				</div>

				{/* Enhanced Popular Skills */}
				<motion.div
					initial={{ opacity: 0, y: 50 }}
					whileInView={{ opacity: 1, y: 0 }}
					transition={{ duration: 0.8, delay: 0.5 }}
					viewport={{ once: true }}
					className="text-center"
				>
					<h3 className="text-3xl md:text-4xl font-bold mb-8 text-stone-900 dark:text-stone-100">
						Trending This Week
					</h3>

					<div className="flex flex-wrap justify-center gap-4 mb-12">
						{[
							'React.js',
							'UI/UX Design',
							'Python',
							'Digital Marketing',
							'Photography',
							'Data Science',
							'Spanish',
							'Video Editing',
							'Machine Learning',
							'Graphic Design'
						].map((skill) => (
							<motion.div
								key={skill}
								whileHover={{ scale: 1.1, y: -2 }}
								whileTap={{ scale: 0.95 }}
							>
								<Badge
									variant="outline"
									className="px-6 py-3 text-lg border-2 hover:bg-sage-50 dark:hover:bg-sage-900 cursor-pointer transition-all duration-300 rounded-full"
								>
									{skill}
								</Badge>
							</motion.div>
						))}
					</div>

					<Button
						size="lg"
						className="bg-gradient-to-r from-sage-600 to-terracotta-600 hover:from-sage-700 hover:to-terracotta-700 text-white px-10 py-6 text-lg rounded-2xl shadow-lg hover:shadow-xl transition-all duration-300 group"
					>
						Explore All Skills
						<ArrowRight className="ml-3 h-5 w-5 group-hover:translate-x-1 transition-transform" />
					</Button>
				</motion.div>
			</div>
		</section>
	);
};

export default SkillsMarketplace;
